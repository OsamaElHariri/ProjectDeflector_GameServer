package gamemechanics

import (
	"errors"
	"math/rand"
)

const (
	UP = iota
	DOWN
	LEFT
	RIGHT
)

const (
	RED_SIDE  = LEFT
	BLUE_SIDE = RIGHT
)

const (
	DESTROY_PAWM = "destroy_pawn"
)

type GameBoardDefenition struct {
	YMax   int
	Events []GameEvent
}

type GameBoard struct {
	Turn       int
	defenition GameBoardDefenition
	XMin       int
	XMax       int
	Pawns      [][]*Pawn
	ScoreBoard ScoreBoard
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

type ScoreBoard struct {
	Red  int
	Blue int
}

func NewGameBoardDefinition() GameBoardDefenition {
	definition := GameBoardDefenition{
		YMax:   5,
		Events: make([]GameEvent, 0),
	}

	return definition
}

func NewGameBoard(defenition GameBoardDefenition) (ProcessedGameBoard, error) {
	gameBoard := GameBoard{
		defenition: NewGameBoardDefinition(),
		Pawns:      make([][]*Pawn, defenition.YMax),
		Turn:       0,
		ScoreBoard: ScoreBoard{},
	}
	gameBoardInProcess := ProcessedGameBoard{
		GameBoard: gameBoard,
	}
	return ProcessEvents(gameBoardInProcess, defenition.Events)
}

func ProcessEvents(gameBoardInProcess ProcessedGameBoard, events []GameEvent) (ProcessedGameBoard, error) {
	for _, event := range events {
		result, err := event.UpdateGameBoard(gameBoardInProcess)
		if err != nil {
			return gameBoardInProcess, err
		}
		gameBoardInProcess = result
	}
	gameBoardInProcess.GameBoard.defenition.Events = append(gameBoardInProcess.GameBoard.defenition.Events, events...)

	return gameBoardInProcess, nil
}

func (gameBoard GameBoard) GetDefenition() GameBoardDefenition {
	return gameBoard.defenition
}

func (gameBoard GameBoard) getPawn(position Position) (*Pawn, error) {
	index, err := getPawnIndex(gameBoard.Pawns, position)
	if err != nil {
		return &Pawn{}, err
	}
	return gameBoard.Pawns[position.Y][index], nil
}

func getPawnIndex(pawns [][]*Pawn, position Position) (int, error) {
	if !isWithinBoard(pawns, position.Y) {
		return 0, errors.New("out of bounds")
	}

	for i, pawn := range pawns[position.Y] {
		if pawn.Position.X == position.X {
			return i, nil
		}
	}

	return 0, errors.New("empty position")
}

func isWithinBoard(pawns [][]*Pawn, yCoord int) bool {
	return yCoord >= 0 && yCoord < len(pawns)
}

func removePawn(pawns [][]*Pawn, position Position) ([][]*Pawn, error) {
	if !isWithinBoard(pawns, position.Y) {
		return pawns, errors.New("invalid pawn position")
	}

	index, err := getPawnIndex(pawns, position)
	if err == nil {
		pawns[position.Y] = append(pawns[position.Y][:index], pawns[position.Y][index+1:]...)
	}

	return pawns, nil
}

func addPawn(pawns [][]*Pawn, newPawn Pawn) ([][]*Pawn, error) {
	if !isWithinBoard(pawns, newPawn.Position.Y) {
		return pawns, errors.New("invalid pawn position")
	}

	index, err := getPawnIndex(pawns, newPawn.Position)
	if err == nil {
		pawns[newPawn.Position.Y][index] = &newPawn
	} else {
		pawns[newPawn.Position.Y] = append(pawns[newPawn.Position.Y], &newPawn)
	}

	return pawns, nil
}

func (gameBoard GameBoard) getNextPawn(currentPosition Position, currentDirection int) (*Pawn, error) {

	if currentDirection == UP {
		for i := currentPosition.Y + 1; i < gameBoard.defenition.YMax; i++ {
			pawn, err := gameBoard.getPawn(position(currentPosition.X, i))
			if err == nil {
				return pawn, nil
			}
		}
		return &Pawn{}, errors.New("no next pawn")
	}

	if currentDirection == DOWN {
		for i := currentPosition.Y - 1; i > 0; i-- {
			pawn, err := gameBoard.getPawn(position(currentPosition.X, i))
			if err == nil {
				return pawn, nil
			}
		}
		return &Pawn{}, errors.New("no next pawn")
	}

	if !isWithinBoard(gameBoard.Pawns, currentPosition.Y) {
		return &Pawn{}, errors.New("invalid current position")
	}

	if currentDirection == LEFT {
		indexNearest := -1
		for index, pawn := range gameBoard.Pawns[currentPosition.Y] {
			if pawn.Position.X < currentPosition.X {
				if indexNearest < 0 || (indexNearest >= 0 && pawn.Position.X > gameBoard.Pawns[currentPosition.Y][indexNearest].Position.X) {
					indexNearest = index
				}
			}
		}
		if indexNearest >= 0 {
			return gameBoard.Pawns[currentPosition.Y][indexNearest], nil
		} else {
			return &Pawn{}, errors.New("no next pawn")
		}
	}

	if currentDirection == RIGHT {
		indexNearest := -1
		for index, pawn := range gameBoard.Pawns[currentPosition.Y] {
			if pawn.Position.X > currentPosition.X {
				if indexNearest < 0 || (indexNearest >= 0 && pawn.Position.X < gameBoard.Pawns[currentPosition.Y][indexNearest].Position.X) {
					indexNearest = index
				}
			}
		}
		if indexNearest >= 0 {
			return gameBoard.Pawns[currentPosition.Y][indexNearest], nil
		} else {
			return &Pawn{}, errors.New("no next pawn")
		}
	}
	return &Pawn{}, errors.New("no next pawn")

}

func ProcessDeflection(gameBoard GameBoard) (GameBoard, []Deflection) {
	currentDirection := UP
	currentPosition := position(0, 0)

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
	if lastDirection == BLUE_SIDE {
		gameBoard.ScoreBoard.Blue += 1
	} else if lastDirection == RED_SIDE {
		gameBoard.ScoreBoard.Red += 1
	}

	return gameBoard, deflections
}

func GetPawnVariants(gameId int, player int, turns int) []string {
	rand.Seed(int64(gameId) + int64(player))

	variants := make([]string, turns)
	for i := 0; i < turns; i++ {
		rand := rand.Float64()
		if rand < 0.5 {
			variants[i] = SLASH
		} else {
			variants[i] = BACKSLASH
		}
	}
	return variants
}

func GetPlayerTurn(turn int) int {
	if turn%2 == 0 {
		return RED_SIDE
	} else {
		return BLUE_SIDE
	}
}

func (gameBoard GameBoard) GetTurnsPlayed(variant string) int {
	count := 0
	for i := 0; i < len(gameBoard.defenition.Events); i++ {
		if gameBoard.defenition.Events[i].DoesConsumeVariant(variant) {
			count += 1
		}
	}
	return count
}
