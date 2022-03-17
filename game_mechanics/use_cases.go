package gamemechanics

import (
	"errors"
	"projectdeflector/game/broadcast"
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

func getProcessedGameBoard(repo repositories.Repository, id string) (ProcessedGameBoard, error) {
	repoDefenition, err := repo.GetGame(id)
	if err != nil {
		return ProcessedGameBoard{}, err
	}

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
	}

	return NewGameBoard(defenition)
}

func (useCase UseCase) GetGame(id string) (ProcessedGameBoard, error) {
	processedGameBoard, err := getProcessedGameBoard(useCase.Repo, id)

	if err != nil {
		return ProcessedGameBoard{}, err
	}

	fireEvent := NewFireDeflectorEvent()
	processedGameBoard, err = ProcessEvents(processedGameBoard, []GameEvent{fireEvent})
	if err != nil {
		return ProcessedGameBoard{}, err
	}

	return processedGameBoard, nil
}

type AddPawnRequest struct {
	X          int
	Y          int
	PlayerSide string
}

type AddPawnResult struct {
	ScoreBoard  map[string]int
	Variants    map[string][]string
	NewPawn     Pawn
	Deflections []Deflection
}

func (res AddPawnResult) ToMap() map[string]interface{} {

	deflections := make([]map[string]interface{}, 0)
	for i := 0; i < len(res.Deflections); i++ {
		deflections = append(deflections, res.Deflections[i].toMap())
	}

	return map[string]interface{}{
		"newPawn":     res.NewPawn.toMap(),
		"deflections": deflections,
	}
}

func (useCase UseCase) AddPawn(gameId string, addPawnRequest AddPawnRequest) (AddPawnResult, error) {
	processedGameBoard, err := getProcessedGameBoard(useCase.Repo, gameId)

	if err != nil {
		return AddPawnResult{}, err
	}

	pawnEvent := NewCreatePawnEvent(NewPosition(addPawnRequest.X, addPawnRequest.Y), addPawnRequest.PlayerSide)
	var newEvents []GameEvent

	newEvents = append(newEvents, pawnEvent)

	processedGameBoard, err = ProcessEvents(processedGameBoard, newEvents)

	if err != nil {
		return AddPawnResult{}, err
	}

	newPawn, err := processedGameBoard.GameBoard.GetPawn(NewPosition(addPawnRequest.X, addPawnRequest.Y))

	if err != nil {
		return AddPawnResult{}, err
	}

	err = useCase.Repo.ReplaceGame(gameId, getInsertDefenition(processedGameBoard.GameBoard.defenition))
	if err != nil {
		return AddPawnResult{}, err
	}

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
	}

	broadcastIds := getBroadcastIds(processedGameBoard, addPawnRequest.PlayerSide)
	broadcast.SocketBroadcast(broadcastIds, "pawn", result.ToMap())

	return result, nil
}

type EndTurnResult struct {
	ScoreBoard        map[string]int
	Variants          map[string][]string
	PlayerTurn        string
	AllDeflections    [][]Deflection
	Winner            string
	MatchPointPlayers map[string]bool
	AvailableShuffles map[string]int
	Deflections       []Deflection
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
		"scoreBoard":        res.ScoreBoard,
		"variants":          res.Variants,
		"playerTurn":        res.PlayerTurn,
		"allDeflections":    allDeflections,
		"winner":            res.Winner,
		"matchPointPlayers": res.MatchPointPlayers,
		"availableShuffles": res.AvailableShuffles,
		"deflections":       deflections,
	}
}

func (useCase UseCase) EndTurn(gameId string, playerSide string) (EndTurnResult, error) {
	processedGameBoard, err := getProcessedGameBoard(useCase.Repo, gameId)

	if err != nil {
		return EndTurnResult{}, err
	}

	allDeflections := make([][]Deflection, 0)

	hasFired := false
	fullOnTurnStart := processedGameBoard.GameBoard.IsFull()

	isDense := true
	for !hasFired || (fullOnTurnStart && isDense) {
		hasFired = true
		fireEvent := NewFireDeflectorEvent()
		processedGameBoard, err = ProcessEvents(processedGameBoard, []GameEvent{fireEvent})
		if err != nil {
			return EndTurnResult{}, err
		}

		if len(processedGameBoard.LastDeflections) > 1 {
			lastDirection := processedGameBoard.LastDeflections[len(processedGameBoard.LastDeflections)-1].ToDirection
			playerId, ok := GetPlayerFromDirection(processedGameBoard.GameBoard.GetDefenition(), lastDirection)

			if ok && processedGameBoard.PlayersInMatchPoint[playerId] {
				winEvent := NewWinEvent(playerId)
				processedGameBoard, err = ProcessEvents(processedGameBoard, []GameEvent{winEvent})

				if err != nil {
					return EndTurnResult{}, err
				}
				break
			}
		}

		allDeflections = append(allDeflections, processedGameBoard.LastDeflections)
		isDense = processedGameBoard.GameBoard.IsDense()
	}

	endTurnEvent := NewEndTurnEvent(playerSide)
	processedGameBoard, err = ProcessEvents(processedGameBoard, []GameEvent{endTurnEvent})

	if err != nil {
		return EndTurnResult{}, err
	}

	if processedGameBoard.GameInProgress {
		matchPointEvents := GetMatchPointEvents(processedGameBoard)
		processedGameBoard, err = ProcessEvents(processedGameBoard, matchPointEvents)

		if err != nil {
			return EndTurnResult{}, err
		}
	}

	err = useCase.Repo.ReplaceGame(gameId, getInsertDefenition(processedGameBoard.GameBoard.defenition))
	if err != nil {
		return EndTurnResult{}, err
	}

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
		ScoreBoard:        processedGameBoard.GameBoard.ScoreBoard,
		Variants:          processedGameBoard.PawnVariants,
		PlayerTurn:        GetPlayerTurn(processedGameBoard.GameBoard),
		Winner:            processedGameBoard.Winner,
		AllDeflections:    allDeflections,
		AvailableShuffles: processedGameBoard.AvailableShuffles,
		MatchPointPlayers: processedGameBoard.PlayersInMatchPoint,
		Deflections:       nextProcessedGameBoard.LastDeflections,
	}

	broadcastIds := getBroadcastIds(processedGameBoard, playerSide)
	broadcast.SocketBroadcast(broadcastIds, "turn", result.ToMap())

	return result, nil
}

type ShuffleResult struct {
	Variants          map[string][]string
	AvailableShuffles map[string]int
}

func (res ShuffleResult) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"variants":          res.Variants,
		"availableShuffles": res.AvailableShuffles,
	}
}

func (useCase UseCase) Shuffle(gameId string, playerSide string) (ShuffleResult, error) {
	processedGameBoard, err := getProcessedGameBoard(useCase.Repo, gameId)

	if err != nil {
		return ShuffleResult{}, err
	}

	skipEvent := NewSkipPawnEvent(playerSide)
	processedGameBoard, err = ProcessEvents(processedGameBoard, []GameEvent{skipEvent})
	if err != nil {
		return ShuffleResult{}, err
	}

	err = useCase.Repo.ReplaceGame(gameId, getInsertDefenition(processedGameBoard.GameBoard.defenition))
	if err != nil {
		return ShuffleResult{}, err
	}

	result := ShuffleResult{
		Variants:          processedGameBoard.PawnVariants,
		AvailableShuffles: processedGameBoard.AvailableShuffles,
	}

	broadcastIds := getBroadcastIds(processedGameBoard, playerSide)
	broadcast.SocketBroadcast(broadcastIds, "shuffle", result.ToMap())

	return result, nil
}

type PeekRequest struct {
	X          int
	Y          int
	PlayerSide string
}

type PeekResult struct {
	NewPawn     Pawn
	Deflections []Deflection
}

func (res PeekResult) ToMap() map[string]interface{} {

	deflections := make([]map[string]interface{}, 0)
	for i := 0; i < len(res.Deflections); i++ {
		deflections = append(deflections, res.Deflections[i].toMap())
	}

	return map[string]interface{}{
		"newPawn":     res.NewPawn.toMap(),
		"deflections": deflections,
	}
}

func (useCase UseCase) Peek(gameId string, peekRequest PeekRequest) (PeekResult, error) {
	processedGameBoard, err := getProcessedGameBoard(useCase.Repo, gameId)

	if err != nil {
		return PeekResult{}, err
	}

	peekPosition := NewPosition(peekRequest.X, peekRequest.Y)
	pawnEvent := NewCreatePawnEvent(peekPosition, peekRequest.PlayerSide)
	fireEvent := NewFireDeflectorEvent()
	var newEvents []GameEvent
	newEvents = append(newEvents, pawnEvent)
	newEvents = append(newEvents, fireEvent)

	processedGameBoard, err = ProcessEvents(processedGameBoard, newEvents)

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
	}

	broadcastIds := getBroadcastIds(processedGameBoard, peekRequest.PlayerSide)
	broadcast.SocketBroadcast(broadcastIds, "peek", result.ToMap())

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
