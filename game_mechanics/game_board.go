package gamemechanics

import (
	"errors"
	"time"
)

const (
	UP = iota
	DOWN
	LEFT
	RIGHT
)

const (
	PLAYER_ONE_SIDE = LEFT
	PLAYER_TWO_SIDE = RIGHT
)

const (
	SET_DURABILITY = "set_durability"
	DESTROY_PAWM   = "destroy_pawn"
)

type GameBoardDefenition struct {
	Id          string
	PlayerIds   []string
	YMax        int
	XMax        int
	Events      []GameEvent
	TargetScore int
	StartTime   int64
	TimePerTurn int
}

type GameBoard struct {
	Turn       int
	defenition GameBoardDefenition
	Pawns      [][]*Pawn
	ScoreBoard map[string]int
}

func (gameBoard GameBoard) toMap() map[string]interface{} {

	pawns := make([][]map[string]interface{}, len(gameBoard.Pawns))
	for i := 0; i < len(gameBoard.Pawns); i++ {
		for j := 0; j < len(gameBoard.Pawns[i]); j++ {
			if gameBoard.Pawns[i][j] != nil {
				pawns[i] = append(pawns[i], gameBoard.Pawns[i][j].toMap())
			} else {
				pawns[i] = append(pawns[i], Pawn{}.toMap())
			}
		}
	}
	return map[string]interface{}{
		"id":          gameBoard.defenition.Id,
		"xMax":        gameBoard.defenition.XMax,
		"xMin":        0,
		"yMax":        gameBoard.defenition.YMax,
		"yMin":        0,
		"turn":        gameBoard.Turn,
		"pawns":       pawns,
		"scoreBoard":  gameBoard.ScoreBoard,
		"timePerTurn": gameBoard.defenition.TimePerTurn,
	}
}

type Deflection struct {
	Position    Position
	ToDirection int
	Events      []DeflectionEvent
}

func (deflection Deflection) toMap() map[string]interface{} {
	events := make([]map[string]interface{}, 0)
	for i := 0; i < len(deflection.Events); i++ {
		events = append(events, deflection.Events[i].toMap())
	}

	return map[string]interface{}{
		"position":    deflection.Position.toMap(),
		"toDirection": deflection.ToDirection,
		"events":      events,
	}
}

type DeflectionEvent struct {
	Name       string
	Position   Position
	Durability int
}

func (deflectionEvent DeflectionEvent) toMap() map[string]interface{} {
	return map[string]interface{}{
		"name":       deflectionEvent.Name,
		"position":   deflectionEvent.Position.toMap(),
		"durability": deflectionEvent.Durability,
	}
}

type DirectedPosition struct {
	Position  Position
	Direction int
}

func NewGameBoardDefinition(gameId string, playerIds []string) GameBoardDefenition {
	definition := GameBoardDefenition{
		PlayerIds:   playerIds,
		Id:          gameId,
		YMax:        2,
		XMax:        2,
		Events:      make([]GameEvent, 0),
		TargetScore: 7,
		StartTime:   time.Now().UnixMilli(),
		TimePerTurn: 45 * 1000,
	}

	return definition
}

func NewGameBoard(defenition GameBoardDefenition) (ProcessedGameBoard, error) {
	return newGameBoard(defenition, RandomVarianceFactory{})
}

func newGameBoard(defenition GameBoardDefenition, varianceFactory VarianceFactory) (ProcessedGameBoard, error) {
	height := defenition.YMax + 1
	width := defenition.XMax + 1

	pawns := make([][]*Pawn, height)

	for i := 0; i < len(pawns); i++ {
		pawns[i] = make([]*Pawn, width)
	}

	scoreBoard := make(map[string]int)
	pawnVariants := make(map[string][]string)
	for _, playerId := range defenition.PlayerIds {
		scoreBoard[playerId] = 0
		pawnVariants[playerId] = varianceFactory.GeneratePawnVariant(getPlayerDigest(defenition, playerId), 1)
	}
	scoreBoard[defenition.PlayerIds[0]] = 1

	events := defenition.Events
	defenition.Events = make([]GameEvent, 0)
	gameBoard := GameBoard{
		defenition: defenition,
		Pawns:      pawns,
		Turn:       0,
		ScoreBoard: scoreBoard,
	}

	playersInMatchPoint := make(map[string]bool)
	availableShuffles := make(map[string]int)
	for _, playerId := range gameBoard.defenition.PlayerIds {
		playersInMatchPoint[playerId] = false
		availableShuffles[playerId] = 1
	}

	gameBoardInProcess := ProcessedGameBoard{
		PlayersInMatchPoint:  playersInMatchPoint,
		AvailableShuffles:    availableShuffles,
		GameBoard:            gameBoard,
		ProcessingEventIndex: 0,
		VarianceFactory:      varianceFactory,
		GameInProgress:       true,
		PawnVariants:         pawnVariants,
		LastTurnEndTime:      defenition.StartTime,
	}
	return ProcessEvents(gameBoardInProcess, events)
}

func ProcessEvents(gameBoardInProcess ProcessedGameBoard, events []GameEvent) (ProcessedGameBoard, error) {
	currentIndex := len(gameBoardInProcess.GameBoard.defenition.Events)
	gameBoardInProcess.GameBoard.defenition.Events = append(gameBoardInProcess.GameBoard.defenition.Events, events...)
	for i, event := range events {
		gameBoardInProcess.ProcessingEventIndex = currentIndex + i
		result, err := event.UpdateGameBoard(gameBoardInProcess)
		if err != nil {
			return gameBoardInProcess, err
		}
		gameBoardInProcess = result
	}

	return gameBoardInProcess, nil
}

func (gameBoard GameBoard) GetDefenition() GameBoardDefenition {
	return gameBoard.defenition
}

func (gameBoard GameBoard) IsFull() bool {
	return gameBoard.getPawnCount() >= gameBoard.getArea()
}

func (gameBoard GameBoard) IsDense() bool {
	return gameBoard.getPawnCount() >= (gameBoard.getArea() - gameBoard.defenition.YMax)
}

func (gameBoard GameBoard) getArea() int {
	return (gameBoard.defenition.XMax + 1) * (gameBoard.defenition.YMax + 1)
}

func (gameBoard GameBoard) CopyScoreBoard() map[string]int {
	currentScoreBoard := map[string]int{}
	for key, value := range gameBoard.ScoreBoard {
		currentScoreBoard[key] = value
	}
	return currentScoreBoard
}

func (gameBoard GameBoard) getPawnCount() int {
	count := 0
	for i := 0; i <= gameBoard.defenition.YMax; i++ {
		for j := 0; j <= gameBoard.defenition.XMax; j++ {
			if gameBoard.Pawns[i][j] != nil {
				count += 1
			}
		}
	}
	return count
}

func (gameBoard GameBoard) GetPawn(position Position) (*Pawn, error) {
	if !isWithinBoard(gameBoard.Pawns, position) {
		return nil, errors.New("invalid pawn position")
	}
	if gameBoard.Pawns[position.Y][position.X] == nil {
		return nil, errors.New("empty pawn position")
	}

	return gameBoard.Pawns[position.Y][position.X], nil
}

func isWithinBoard(pawns [][]*Pawn, position Position) bool {
	height := len(pawns)
	width := len(pawns[0])
	return position.X >= 0 && position.Y >= 0 && position.X < width && position.Y < height
}

func removePawn(pawns [][]*Pawn, position Position) ([][]*Pawn, error) {
	if !isWithinBoard(pawns, position) {
		return pawns, errors.New("invalid pawn position")
	}
	pawns[position.Y][position.X] = nil

	return pawns, nil
}

func addPawn(pawns [][]*Pawn, newPawn Pawn) ([][]*Pawn, error) {
	if !isWithinBoard(pawns, newPawn.Position) {
		return pawns, errors.New("invalid pawn position")
	}

	pawns[newPawn.Position.Y][newPawn.Position.X] = &newPawn
	return pawns, nil
}

func (gameBoard GameBoard) getNextPawn(currentPosition Position, currentDirection int) (*Pawn, error) {

	if currentDirection == UP {
		for i := currentPosition.Y + 1; i <= gameBoard.defenition.YMax; i++ {
			pawn, err := gameBoard.GetPawn(position(currentPosition.X, i))
			if pawn != nil && err == nil {
				return pawn, nil
			}
		}
		return &Pawn{}, errors.New("no next pawn")
	}

	if currentDirection == DOWN {
		for i := currentPosition.Y - 1; i >= 0; i-- {
			pawn, err := gameBoard.GetPawn(position(currentPosition.X, i))
			if pawn != nil && err == nil {
				return pawn, nil
			}
		}
		return &Pawn{}, errors.New("no next pawn")
	}

	if currentDirection == RIGHT {
		for i := currentPosition.X + 1; i <= gameBoard.defenition.XMax; i++ {
			pawn, err := gameBoard.GetPawn(position(i, currentPosition.Y))
			if pawn != nil && err == nil {
				return pawn, nil
			}
		}
		return &Pawn{}, errors.New("no next pawn")
	}

	if currentDirection == LEFT {
		for i := currentPosition.X - 1; i >= 0; i-- {
			pawn, err := gameBoard.GetPawn(position(i, currentPosition.Y))
			if pawn != nil && err == nil {
				return pawn, nil
			}
		}
		return &Pawn{}, errors.New("no next pawn")
	}

	return &Pawn{}, errors.New("no next pawn")
}

func getEdgeDeflection(gameBoard GameBoard, lastDeflection Deflection) Deflection {
	var pos Position
	if lastDeflection.ToDirection == UP {
		pos = position(lastDeflection.Position.X, gameBoard.defenition.YMax+1)
	} else if lastDeflection.ToDirection == DOWN {
		pos = position(lastDeflection.Position.X, -1)
	} else if lastDeflection.ToDirection == LEFT {
		pos = position(-1, lastDeflection.Position.Y)
	} else if lastDeflection.ToDirection == RIGHT {
		pos = position(gameBoard.defenition.XMax+1, lastDeflection.Position.Y)
	}

	return Deflection{
		ToDirection: lastDeflection.ToDirection,
		Position:    pos,
	}
}

func ProcessDeflection(gameBoard GameBoard, current DirectedPosition) (GameBoard, []Deflection) {
	currentPosition, currentDirection := current.Position, current.Direction
	deflections := []Deflection{
		{
			Position:    currentPosition,
			ToDirection: currentDirection,
			Events:      make([]DeflectionEvent, 0),
		},
	}

	for {
		pawn, err := gameBoard.getNextPawn(currentPosition, currentDirection)
		if err != nil {
			break
		}
		currentPosition = pawn.Position
		currentDirection = pawn.getDeflectedDirection(currentDirection)
		pawn.Durability -= 1
		events := make([]DeflectionEvent, 0)
		events = append(events, DeflectionEvent{
			Name:       SET_DURABILITY,
			Position:   pawn.Position,
			Durability: pawn.Durability,
		})

		if pawn.Durability == 0 {
			events = append(events, DeflectionEvent{
				Name:     DESTROY_PAWM,
				Position: pawn.Position,
			})

			gameBoard.Pawns, err = removePawn(gameBoard.Pawns, pawn.Position)
			if err != nil {
				return gameBoard, deflections
			}
		}

		deflections = append(deflections, Deflection{
			Position:    currentPosition,
			ToDirection: currentDirection,
			Events:      events,
		})
	}

	lastDeflection := deflections[len(deflections)-1]
	deflections = append(deflections, getEdgeDeflection(gameBoard, lastDeflection))

	playerId, ok := GetPlayerFromDirection(gameBoard.defenition, lastDeflection.ToDirection)
	if ok {
		gameBoard.ScoreBoard[playerId] += 1
	}

	return gameBoard, deflections
}

func GetPlayerFromDirection(defenition GameBoardDefenition, direction int) (string, bool) {
	if direction == LEFT {
		return defenition.PlayerIds[0], true
	} else if direction == RIGHT {
		return defenition.PlayerIds[1], true
	}
	return "", false
}

func GetPlayerTurn(gameBoard GameBoard) string {
	return gameBoard.defenition.PlayerIds[gameBoard.Turn%len(gameBoard.defenition.PlayerIds)]
}

func getPlayerDigest(defenition GameBoardDefenition, playerId string) string {
	return defenition.Id + playerId
}

func GetMatchPointEvents(gameBoardInPrccess ProcessedGameBoard) []GameEvent {
	matchPointEvents := make([]GameEvent, 0)
	for _, playerId := range gameBoardInPrccess.GameBoard.defenition.PlayerIds {
		if gameBoardInPrccess.GameBoard.ScoreBoard[playerId] >= gameBoardInPrccess.GameBoard.defenition.TargetScore && !gameBoardInPrccess.PlayersInMatchPoint[playerId] {
			matchPointEvents = append(matchPointEvents, NewMatchPointEvent(playerId))
		}
	}
	return matchPointEvents
}
