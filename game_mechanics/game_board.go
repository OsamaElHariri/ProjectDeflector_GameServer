package gamemechanics

import (
	"errors"
	"strconv"
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
	DESTROY_PAWM = "destroy_pawn"
)

type GameBoardDefenition struct {
	PlayerIds   []string
	Id          int
	YMax        int
	XMax        int
	Events      []GameEvent
	TargetScore int
}

type GameBoard struct {
	Turn       int
	defenition GameBoardDefenition
	Pawns      [][]*Pawn
	ScoreBoard map[string]int
}

type Deflection struct {
	Position    Position
	ToDirection int
	Events      []DeflectionEvent
}

type DeflectionEvent struct {
	Name     string
	Position Position
}

type DirectedPosition struct {
	Position  Position
	Direction int
}

func NewGameBoardDefinition(gameId int) GameBoardDefenition {
	definition := GameBoardDefenition{
		PlayerIds:   []string{"red", "blue"},
		Id:          gameId,
		YMax:        2,
		XMax:        2,
		Events:      make([]GameEvent, 0),
		TargetScore: 7,
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
	for index, playerId := range defenition.PlayerIds {
		scoreBoard[playerId] = index + 1
	}
	gameBoard := GameBoard{
		defenition: NewGameBoardDefinition(defenition.Id),
		Pawns:      pawns,
		Turn:       0,
		ScoreBoard: scoreBoard,
	}

	playersInMatchPoint := make(map[string]bool)
	for _, playerId := range gameBoard.defenition.PlayerIds {
		playersInMatchPoint[playerId] = false
	}

	gameBoardInProcess := ProcessedGameBoard{
		PlayersInMatchPoint:  playersInMatchPoint,
		GameBoard:            gameBoard,
		ProcessingEventIndex: 0,
		VarianceFactory:      varianceFactory,
	}
	return ProcessEvents(gameBoardInProcess, defenition.Events)
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

func (gameBoard GameBoard) getPawn(position Position) (*Pawn, error) {
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
			pawn, err := gameBoard.getPawn(position(currentPosition.X, i))
			if pawn != nil && err == nil {
				return pawn, nil
			}
		}
		return &Pawn{}, errors.New("no next pawn")
	}

	if currentDirection == DOWN {
		for i := currentPosition.Y - 1; i >= 0; i-- {
			pawn, err := gameBoard.getPawn(position(currentPosition.X, i))
			if pawn != nil && err == nil {
				return pawn, nil
			}
		}
		return &Pawn{}, errors.New("no next pawn")
	}

	if currentDirection == RIGHT {
		for i := currentPosition.X + 1; i <= gameBoard.defenition.XMax; i++ {
			pawn, err := gameBoard.getPawn(position(i, currentPosition.Y))
			if pawn != nil && err == nil {
				return pawn, nil
			}
		}
		return &Pawn{}, errors.New("no next pawn")
	}

	if currentDirection == LEFT {
		for i := currentPosition.X - 1; i >= 0; i-- {
			pawn, err := gameBoard.getPawn(position(i, currentPosition.Y))
			if pawn != nil && err == nil {
				return pawn, nil
			}
		}
		return &Pawn{}, errors.New("no next pawn")
	}

	return &Pawn{}, errors.New("no next pawn")

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

	// being stuck in an infinite loop is not possible when given valid inputs.
	// an upperbound of (100 + gameBoard.Turn)*2 protects against the possibility of
	// an infinite loop in case unexpected inputs are passed in
	for i := 0; i < (100+gameBoard.Turn)*2; i++ {
		pawn, err := gameBoard.getNextPawn(currentPosition, currentDirection)
		if err != nil {
			break
		}
		currentPosition = pawn.Position
		currentDirection = pawn.getDeflectedDirection(currentDirection)
		pawn.Durability -= 1
		events := make([]DeflectionEvent, 0)

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

	lastDirection := deflections[len(deflections)-1].ToDirection
	if lastDirection == PLAYER_ONE_SIDE {
		gameBoard.ScoreBoard[gameBoard.defenition.PlayerIds[0]] += 1
	} else if lastDirection == PLAYER_TWO_SIDE {
		gameBoard.ScoreBoard[gameBoard.defenition.PlayerIds[1]] += 1
	}

	return gameBoard, deflections
}

func GetPlayerTurn(gameBoard GameBoard) string {
	return gameBoard.defenition.PlayerIds[gameBoard.Turn%len(gameBoard.defenition.PlayerIds)]
}

func (gameBoard GameBoard) GetTurnsPlayed(variant string) int {
	return getTurnsPlayed(gameBoard.defenition.Events, variant)
}

func getTurnsPlayed(events []GameEvent, variant string) int {
	count := 0
	for i := 0; i < len(events); i++ {
		if events[i].DoesConsumeVariant(variant) {
			count += 1
		}
	}
	return count
}

func (gameBoard GameBoard) GetPlayerDigest(playerId string) string {
	return strconv.Itoa(gameBoard.defenition.Id) + playerId
}

func GetMatchPointEvents(gameBoardInPrccess ProcessedGameBoard) []GameEvent {
	matchPointEvents := make([]GameEvent, 0)
	for _, playerId := range gameBoardInPrccess.GameBoard.defenition.PlayerIds {
		if gameBoardInPrccess.GameBoard.ScoreBoard[playerId] > gameBoardInPrccess.GameBoard.defenition.TargetScore && !gameBoardInPrccess.PlayersInMatchPoint[playerId] {
			matchPointEvents = append(matchPointEvents, NewMatchPointEvent(playerId))
		}
	}
	return matchPointEvents
}

func GetPawnVariants(gameBoardInPrccess ProcessedGameBoard) map[string][]string {
	variants := make(map[string][]string)
	for _, playerId := range gameBoardInPrccess.GameBoard.defenition.PlayerIds {
		variants[playerId] = gameBoardInPrccess.VarianceFactory.GeneratePawnVariant(gameBoardInPrccess.GameBoard.GetPlayerDigest(playerId), gameBoardInPrccess.GameBoard.GetTurnsPlayed("blue")+2)
	}
	return variants
}
