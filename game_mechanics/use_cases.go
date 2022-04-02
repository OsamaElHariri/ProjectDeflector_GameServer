package gamemechanics

import (
	"errors"
	"projectdeflector/game/network"
	"projectdeflector/game/repositories"
)

type UseCase struct {
	Repo repositories.Repository
}

func getInsertDefenition(defenition GameBoardDefenition) repositories.InserGameBoardDefenition {
	mappedEvents := make([]map[string]interface{}, 0)
	for i := 0; i < len(defenition.Events); i++ {
		mappedEvents = append(mappedEvents, defenition.Events[i].Encode())
	}

	return repositories.InserGameBoardDefenition{
		PlayerIds:   defenition.PlayerIds,
		XMax:        defenition.XMax,
		YMax:        defenition.YMax,
		TargetScore: defenition.TargetScore,
		Events:      mappedEvents,
		TimePerTurn: defenition.TimePerTurn,
		StartTime:   defenition.StartTime,
	}
}

func (useCase UseCase) CreateNewGame(playerIds []string) (string, error) {

	if len(playerIds) != 2 {
		return "", errors.New("a game can only have two players")
	}

	defenition := NewGameBoardDefinition("test", playerIds)

	insert := getInsertDefenition(defenition)
	return useCase.Repo.InsertGame(insert)
}

func getLockedProcessedGameBoard(repo repositories.Repository, id string) (ProcessedGameBoard, error) {
	repoDefenition, err := repo.GetGameAndLock(id)
	if err != nil {
		return ProcessedGameBoard{}, err
	}
	return getGameBoardFromDbDefenition(repoDefenition)
}

func getProcessedGameBoard(repo repositories.Repository, id string) (ProcessedGameBoard, error) {
	repoDefenition, err := repo.GetGame(id)
	if err != nil {
		return ProcessedGameBoard{}, err
	}
	return getGameBoardFromDbDefenition(repoDefenition)
}

func getGameBoardFromDbDefenition(repoDefenition repositories.GetGameBoardDefenitionResult) (ProcessedGameBoard, error) {
	decodedEvents := make([]GameEvent, 0)
	for i := 0; i < len(repoDefenition.Events); i++ {

		event, err := DecodeGameEvent(repoDefenition.Events[i])
		if err != nil {
			return ProcessedGameBoard{}, err
		}
		decodedEvents = append(decodedEvents, event)
	}

	defenition := GameBoardDefenition{
		Id:          repoDefenition.Id,
		PlayerIds:   repoDefenition.PlayerIds,
		Events:      decodedEvents,
		YMax:        repoDefenition.YMax,
		XMax:        repoDefenition.XMax,
		TargetScore: repoDefenition.TargetScore,
		StartTime:   repoDefenition.StartTime,
		TimePerTurn: repoDefenition.TimePerTurn,
	}

	return NewGameBoard(defenition)
}

type GetGameResult struct {
	ProcessedGameBoard     ProcessedGameBoard
	NextProcessedGameBoard ProcessedGameBoard
	EventCount             int
}

func (res GetGameResult) ToMap() map[string]interface{} {
	res.ProcessedGameBoard.LastDeflections = res.NextProcessedGameBoard.LastDeflections
	toMap := res.ProcessedGameBoard.toMap()
	toMap["postDeflectionPartialGameBoard"] = PostDeflectionPartialGameBoard{
		PreviousScoreBoard: res.ProcessedGameBoard.GameBoard.ScoreBoard,
		ScoreBoard:         res.NextProcessedGameBoard.GameBoard.ScoreBoard,
	}
	toMap["eventCount"] = res.EventCount
	return toMap
}

func (useCase UseCase) GetGame(id string) (GetGameResult, error) {
	processedGameBoard, err := getProcessedGameBoard(useCase.Repo, id)

	if err != nil {
		return GetGameResult{}, err
	}
	eventCount := len(processedGameBoard.GameBoard.defenition.Events)

	nextProcessedGameBoard, err := NewGameBoard(processedGameBoard.GameBoard.GetDefenition())
	if err != nil {
		return GetGameResult{}, err
	}
	fireEvent := NewFireDeflectorEvent()
	nextProcessedGameBoard, err = ProcessEvents(nextProcessedGameBoard, []GameEvent{fireEvent})
	if err != nil {
		return GetGameResult{}, err
	}

	return GetGameResult{
		ProcessedGameBoard:     processedGameBoard,
		EventCount:             eventCount,
		NextProcessedGameBoard: nextProcessedGameBoard,
	}, nil
}

func (useCase UseCase) GetOngoingGameId(playerId string) (string, error) {
	dbGameBoard, err := useCase.Repo.GetOngoingPlayerGame(playerId)
	if err != nil {
		return "", err
	}
	return dbGameBoard.Id, nil
}

type AddPawnRequest struct {
	X          int
	Y          int
	PlayerSide string
}

type AddPawnResult struct {
	ScoreBoard                     map[string]int
	Variants                       map[string][]string
	NewPawn                        Pawn
	Deflections                    []Deflection
	PostDeflectionPartialGameBoard PostDeflectionPartialGameBoard
	EventCount                     int
	PreviousEventCount             int
}

func (res AddPawnResult) ToMap() map[string]interface{} {

	deflections := make([]map[string]interface{}, 0)
	for i := 0; i < len(res.Deflections); i++ {
		deflections = append(deflections, res.Deflections[i].toMap())
	}

	return map[string]interface{}{
		"newPawn":                        res.NewPawn.toMap(),
		"deflections":                    deflections,
		"postDeflectionPartialGameBoard": res.PostDeflectionPartialGameBoard,
		"variants":                       res.Variants,
		"scoreBoard":                     res.ScoreBoard,
		"eventCount":                     res.EventCount,
		"previousEventCount":             res.PreviousEventCount,
	}
}

func (useCase UseCase) AddPawn(gameId string, addPawnRequest AddPawnRequest) (AddPawnResult, error) {
	processedGameBoard, err := getLockedProcessedGameBoard(useCase.Repo, gameId)

	if err != nil {
		return AddPawnResult{}, err
	}
	previousEventCount := len(processedGameBoard.GameBoard.defenition.Events)

	pawnEvent := NewCreatePawnEvent(NewPosition(addPawnRequest.X, addPawnRequest.Y), addPawnRequest.PlayerSide)
	var newEvents []GameEvent

	newEvents = append(newEvents, pawnEvent)

	processedGameBoard, err = ProcessEvents(processedGameBoard, newEvents)

	if err != nil {
		useCase.Repo.UnlockGame(gameId)
		return AddPawnResult{}, err
	}

	newPawn, err := processedGameBoard.GameBoard.GetPawn(NewPosition(addPawnRequest.X, addPawnRequest.Y))

	if err != nil {
		useCase.Repo.UnlockGame(gameId)
		return AddPawnResult{}, err
	}

	err = useCase.Repo.ReplaceGame(gameId, getInsertDefenition(processedGameBoard.GameBoard.defenition))
	if err != nil {
		useCase.Repo.UnlockGame(gameId)
		return AddPawnResult{}, err
	}
	eventCount := len(processedGameBoard.GameBoard.defenition.Events)

	nextProcessedGameBoard, err := NewGameBoard(processedGameBoard.GameBoard.GetDefenition())
	if err != nil {
		return AddPawnResult{}, err
	}

	fireEvent := NewFireDeflectorEvent()
	nextProcessedGameBoard, err = ProcessEvents(nextProcessedGameBoard, []GameEvent{fireEvent})
	if err != nil {
		return AddPawnResult{}, err
	}

	result := AddPawnResult{
		ScoreBoard:  processedGameBoard.GameBoard.ScoreBoard,
		Variants:    processedGameBoard.PawnVariants,
		NewPawn:     *newPawn,
		Deflections: nextProcessedGameBoard.LastDeflections,
		PostDeflectionPartialGameBoard: PostDeflectionPartialGameBoard{
			PreviousScoreBoard: processedGameBoard.GameBoard.ScoreBoard,
			ScoreBoard:         nextProcessedGameBoard.GameBoard.ScoreBoard,
		},
		EventCount:         eventCount,
		PreviousEventCount: previousEventCount,
	}

	broadcastIds := getBroadcastIds(processedGameBoard, addPawnRequest.PlayerSide)
	network.SocketBroadcast(broadcastIds, "pawn", result.ToMap())

	return result, nil
}

type PostDeflectionPartialGameBoard struct {
	PreviousScoreBoard map[string]int `json:"previousScoreBoard"`
	ScoreBoard         map[string]int `json:"scoreBoard"`
}

type EndTurnResult struct {
	ScoreBoard                         map[string]int
	Variants                           map[string][]string
	PlayerTurn                         string
	AllDeflections                     [][]Deflection
	AllPostDeflectionPartialGameBoards []PostDeflectionPartialGameBoard
	Winner                             string
	MatchPointPlayers                  map[string]bool
	AvailableShuffles                  map[string]int
	Deflections                        []Deflection
	PostDeflectionPartialGameBoard     PostDeflectionPartialGameBoard
	LastTurnEndTime                    int64
	EventCount                         int
	PreviousEventCount                 int
}

func (res EndTurnResult) ToMap() map[string]interface{} {
	allDeflections := make([][]map[string]interface{}, 0)
	for i := 0; i < len(res.AllDeflections); i++ {
		deflections := make([]map[string]interface{}, 0)
		for j := 0; j < len(res.AllDeflections[i]); j++ {
			deflections = append(deflections, res.AllDeflections[i][j].toMap())
		}
		allDeflections = append(allDeflections, deflections)
	}

	deflections := make([]map[string]interface{}, 0)
	for i := 0; i < len(res.Deflections); i++ {
		deflections = append(deflections, res.Deflections[i].toMap())
	}

	return map[string]interface{}{
		"scoreBoard":                         res.ScoreBoard,
		"variants":                           res.Variants,
		"playerTurn":                         res.PlayerTurn,
		"allDeflections":                     allDeflections,
		"winner":                             res.Winner,
		"matchPointPlayers":                  res.MatchPointPlayers,
		"availableShuffles":                  res.AvailableShuffles,
		"deflections":                        deflections,
		"eventCount":                         res.EventCount,
		"previousEventCount":                 res.PreviousEventCount,
		"lastTurnEndTime":                    res.LastTurnEndTime,
		"allPostDeflectionPartialGameBoards": res.AllPostDeflectionPartialGameBoards,
		"postDeflectionPartialGameBoard":     res.PostDeflectionPartialGameBoard,
	}
}

func (useCase UseCase) EndTurn(gameId string, playerSide string) (EndTurnResult, error) {
	processedGameBoard, err := getLockedProcessedGameBoard(useCase.Repo, gameId)

	if err != nil {
		return EndTurnResult{}, err
	}
	endResult, err := endGameTurn(useCase.Repo, processedGameBoard, playerSide)
	if err != nil {
		return EndTurnResult{}, err
	}
	broadcastIds := getBroadcastIds(processedGameBoard, playerSide)
	network.SocketBroadcast(broadcastIds, "turn", endResult.ToMap())

	if processedGameBoard.Winner != "" {
		notifyUserServiceOfGameEnd(useCase.Repo, processedGameBoard)
	}

	return endResult, nil
}

func (useCase UseCase) ExpireTurn(gameId string, playerSide string, eventCount int) (EndTurnResult, error) {
	processedGameBoard, err := getLockedProcessedGameBoard(useCase.Repo, gameId)

	if err != nil {
		return EndTurnResult{}, err
	}

	if len(processedGameBoard.GameBoard.defenition.Events) != eventCount {
		return EndTurnResult{}, errors.New("turn already ended")
	}

	endResult, err := endGameTurn(useCase.Repo, processedGameBoard, "system")
	if err != nil {
		return EndTurnResult{}, err
	}

	network.SocketBroadcast(processedGameBoard.GameBoard.defenition.PlayerIds, "turn", endResult.ToMap())

	if endResult.Winner != "" {
		notifyUserServiceOfGameEnd(useCase.Repo, processedGameBoard)
	}

	return endResult, nil
}

func endGameTurn(repo repositories.Repository, processedGameBoard ProcessedGameBoard, playerSide string) (EndTurnResult, error) {
	gameId := processedGameBoard.GameBoard.defenition.Id
	previousEventCount := len(processedGameBoard.GameBoard.defenition.Events)
	allDeflections := make([][]Deflection, 0)
	partialGameBoards := make([]PostDeflectionPartialGameBoard, 0)

	hasFired := false
	fullOnTurnStart := processedGameBoard.GameBoard.IsFull()

	isDense := true
	for !hasFired || (fullOnTurnStart && isDense) {
		hasFired = true
		scoreBoard := processedGameBoard.GameBoard.CopyScoreBoard()
		fireEvent := NewFireDeflectorEvent()
		var err error
		processedGameBoard, err = ProcessEvents(processedGameBoard, []GameEvent{fireEvent})
		if err != nil {
			repo.UnlockGame(gameId)
			return EndTurnResult{}, err
		}

		if len(processedGameBoard.LastDeflections) > 1 {
			lastDirection := processedGameBoard.LastDeflections[len(processedGameBoard.LastDeflections)-1].ToDirection

			partialGameBoards = append(partialGameBoards, PostDeflectionPartialGameBoard{
				PreviousScoreBoard: scoreBoard,
				ScoreBoard:         processedGameBoard.GameBoard.CopyScoreBoard(),
			})
			allDeflections = append(allDeflections, processedGameBoard.LastDeflections)
			isDense = processedGameBoard.GameBoard.IsDense()

			playerId, ok := GetPlayerFromDirection(processedGameBoard.GameBoard.GetDefenition(), lastDirection)

			if ok && processedGameBoard.PlayersInMatchPoint[playerId] {
				winEvent := NewWinEvent(playerId)
				processedGameBoard, err = ProcessEvents(processedGameBoard, []GameEvent{winEvent})

				if err != nil {
					repo.UnlockGame(gameId)
					return EndTurnResult{}, err
				}
				break
			}
		}
	}

	endTurnEvent := NewEndTurnEvent(playerSide)
	processedGameBoard, err := ProcessEvents(processedGameBoard, []GameEvent{endTurnEvent})

	if err != nil {
		repo.UnlockGame(gameId)
		return EndTurnResult{}, err
	}

	if processedGameBoard.GameInProgress {
		matchPointEvents := GetMatchPointEvents(processedGameBoard)
		processedGameBoard, err = ProcessEvents(processedGameBoard, matchPointEvents)

		if err != nil {
			repo.UnlockGame(gameId)
			return EndTurnResult{}, err
		}
	}

	insert := getInsertDefenition(processedGameBoard.GameBoard.defenition)
	insert.Winner = processedGameBoard.Winner

	err = repo.ReplaceGame(gameId, insert)
	if err != nil {
		repo.UnlockGame(gameId)
		return EndTurnResult{}, err
	}
	eventCount := len(processedGameBoard.GameBoard.defenition.Events)

	nextProcessedGameBoard, err := NewGameBoard(processedGameBoard.GameBoard.GetDefenition())
	if err != nil {
		return EndTurnResult{}, err
	}
	fireEvent := NewFireDeflectorEvent()
	nextProcessedGameBoard, err = ProcessEvents(nextProcessedGameBoard, []GameEvent{fireEvent})

	if err != nil {
		return EndTurnResult{}, err
	}

	result := EndTurnResult{
		ScoreBoard:                         processedGameBoard.GameBoard.ScoreBoard,
		Variants:                           processedGameBoard.PawnVariants,
		PlayerTurn:                         GetPlayerTurn(processedGameBoard.GameBoard),
		Winner:                             processedGameBoard.Winner,
		AllDeflections:                     allDeflections,
		AllPostDeflectionPartialGameBoards: partialGameBoards,
		AvailableShuffles:                  processedGameBoard.AvailableShuffles,
		MatchPointPlayers:                  processedGameBoard.PlayersInMatchPoint,
		Deflections:                        nextProcessedGameBoard.LastDeflections,
		PostDeflectionPartialGameBoard: PostDeflectionPartialGameBoard{
			PreviousScoreBoard: processedGameBoard.GameBoard.ScoreBoard,
			ScoreBoard:         nextProcessedGameBoard.GameBoard.ScoreBoard,
		},
		EventCount:         eventCount,
		PreviousEventCount: previousEventCount,
		LastTurnEndTime:    processedGameBoard.LastTurnEndTime,
	}
	return result, nil
}

func notifyUserServiceOfGameEnd(repo repositories.Repository, processedGameBoard ProcessedGameBoard) {
	repoStatUpdates, err := repo.GetPlayersGameStats(processedGameBoard.GameBoard.defenition.PlayerIds)
	statUpdates := []network.GameEndUserUpdate{}
	for i := 0; i < len(repoStatUpdates); i++ {
		statUpdates = append(statUpdates, network.GameEndUserUpdate{
			PlayerId: repoStatUpdates[i].PlayerId,
			Games:    repoStatUpdates[i].Games,
			Wins:     repoStatUpdates[i].Wins,
		})
	}

	if err == nil {
		network.NotifyUserServiceGameEnd(statUpdates)
	}
}

type ShuffleResult struct {
	Variants           map[string][]string
	AvailableShuffles  map[string]int
	EventCount         int
	PreviousEventCount int
}

func (res ShuffleResult) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"variants":           res.Variants,
		"availableShuffles":  res.AvailableShuffles,
		"eventCount":         res.EventCount,
		"previousEventCount": res.PreviousEventCount,
	}
}

func (useCase UseCase) Shuffle(gameId string, playerSide string) (ShuffleResult, error) {
	processedGameBoard, err := getLockedProcessedGameBoard(useCase.Repo, gameId)

	if err != nil {
		return ShuffleResult{}, err
	}
	previousEventCount := len(processedGameBoard.GameBoard.defenition.Events)

	skipEvent := NewSkipPawnEvent(playerSide)
	processedGameBoard, err = ProcessEvents(processedGameBoard, []GameEvent{skipEvent})
	if err != nil {
		useCase.Repo.UnlockGame(gameId)
		return ShuffleResult{}, err
	}

	err = useCase.Repo.ReplaceGame(gameId, getInsertDefenition(processedGameBoard.GameBoard.defenition))
	if err != nil {
		useCase.Repo.UnlockGame(gameId)
		return ShuffleResult{}, err
	}
	eventCount := len(processedGameBoard.GameBoard.defenition.Events)

	result := ShuffleResult{
		Variants:           processedGameBoard.PawnVariants,
		AvailableShuffles:  processedGameBoard.AvailableShuffles,
		EventCount:         eventCount,
		PreviousEventCount: previousEventCount,
	}

	broadcastIds := getBroadcastIds(processedGameBoard, playerSide)
	network.SocketBroadcast(broadcastIds, "shuffle", result.ToMap())

	return result, nil
}

type PeekRequest struct {
	X          int
	Y          int
	PlayerSide string
}

type PeekResult struct {
	NewPawn                        Pawn
	Deflections                    []Deflection
	EventCount                     int
	PreviousEventCount             int
	PostDeflectionPartialGameBoard PostDeflectionPartialGameBoard
}

func (res PeekResult) ToMap() map[string]interface{} {

	deflections := make([]map[string]interface{}, 0)
	for i := 0; i < len(res.Deflections); i++ {
		deflections = append(deflections, res.Deflections[i].toMap())
	}

	return map[string]interface{}{
		"newPawn":                        res.NewPawn.toMap(),
		"deflections":                    deflections,
		"eventCount":                     res.EventCount,
		"previousEventCount":             res.PreviousEventCount,
		"postDeflectionPartialGameBoard": res.PostDeflectionPartialGameBoard,
	}
}

func (useCase UseCase) Peek(gameId string, peekRequest PeekRequest) (PeekResult, error) {
	processedGameBoard, err := getProcessedGameBoard(useCase.Repo, gameId)

	if err != nil {
		return PeekResult{}, err
	}
	previousEventCount := len(processedGameBoard.GameBoard.defenition.Events)

	peekPosition := NewPosition(peekRequest.X, peekRequest.Y)
	pawnEvent := NewCreatePawnEvent(peekPosition, peekRequest.PlayerSide)

	processedGameBoard, err = ProcessEvents(processedGameBoard, []GameEvent{pawnEvent})

	if err != nil {
		return PeekResult{}, err
	}
	scoreBoard := processedGameBoard.GameBoard.CopyScoreBoard()

	fireEvent := NewFireDeflectorEvent()
	processedGameBoard, err = ProcessEvents(processedGameBoard, []GameEvent{fireEvent})

	if err != nil {
		return PeekResult{}, err
	}

	newPawn, err := processedGameBoard.GameBoard.GetPawn(NewPosition(peekRequest.X, peekRequest.Y))
	if err != nil {
		return PeekResult{}, err
	}

	result := PeekResult{
		NewPawn:     *newPawn,
		Deflections: processedGameBoard.LastDeflections,
		PostDeflectionPartialGameBoard: PostDeflectionPartialGameBoard{
			PreviousScoreBoard: scoreBoard,
			ScoreBoard:         processedGameBoard.GameBoard.ScoreBoard,
		},
		EventCount:         previousEventCount,
		PreviousEventCount: previousEventCount,
	}

	broadcastIds := getBroadcastIds(processedGameBoard, peekRequest.PlayerSide)
	network.SocketBroadcast(broadcastIds, "peek", result.ToMap())

	return result, nil
}

func getBroadcastIds(processedGameBoard ProcessedGameBoard, currentPlayer string) []string {
	broadcastIds := make([]string, 0)
	for i := 0; i < len(processedGameBoard.GameBoard.GetDefenition().PlayerIds); i++ {
		id := processedGameBoard.GameBoard.GetDefenition().PlayerIds[i]
		if id != currentPlayer {
			broadcastIds = append(broadcastIds, id)
		}
	}
	return broadcastIds
}
