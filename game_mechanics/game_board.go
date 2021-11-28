package gamemechanics

import (
	"errors"
)

const (
	UP = iota
	DOWN
	LEFT
	RIGHT
)

type GameBoardDefenition struct {
	YMax   int
	Events []GameEvent
}

type GameBoard struct {
	defenition GameBoardDefenition
	XMin       int
	XMax       int
	Pawns      [][]Pawn
}

type Deflection struct {
	Position    Position
	ToDirection int
}

func NewGameBoardDefinition() GameBoardDefenition {
	definition := GameBoardDefenition{
		YMax:   5,
		Events: []GameEvent{},
	}

	return definition
}

func NewGameBoard(defenition GameBoardDefenition) (GameBoard, error) {
	gameBoard := GameBoard{
		defenition: defenition,
		Pawns:      make([][]Pawn, defenition.YMax),
	}
	for _, event := range defenition.Events {
		gameBoard = ApplyEvent(gameBoard, event)
	}

	return gameBoard, nil
}

func (gameBoard GameBoard) getPawn(position Position) (Pawn, error) {
	index, err := getPawnIndex(gameBoard.Pawns, position)
	if err != nil {
		return Pawn{}, err
	}
	return gameBoard.Pawns[position.Y][index], nil
}

func getPawnIndex(pawns [][]Pawn, position Position) (int, error) {
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

func isWithinBoard(pawns [][]Pawn, yCoord int) bool {
	return yCoord >= 0 && yCoord < len(pawns)
}

func ApplyEvent(gameBoard GameBoard, event GameEvent) GameBoard {
	if event.name == CREATE_PAWN {
		updatedPawns, err := addPawn(gameBoard.Pawns, event)
		if err == nil {
			gameBoard.Pawns = updatedPawns

		}
	}
	return gameBoard
}

func addPawn(pawns [][]Pawn, event GameEvent) ([][]Pawn, error) {
	if !isWithinBoard(pawns, event.position.Y) {
		return pawns, errors.New("invalid pawn position")
	}

	newPawn := Pawn{
		Position:   event.position,
		Name:       event.targetType,
		TurnPlaced: 0,
	}

	index, err := getPawnIndex(pawns, event.position)
	if err == nil {
		pawns[event.position.Y][index] = newPawn
	} else {
		pawns[event.position.Y] = append(pawns[event.position.Y], newPawn)
	}

	return pawns, nil
}

func (gameBoard GameBoard) getNextPawn(currentPosition Position, currentDirection int) (Pawn, error) {

	if currentDirection == UP {
		for i := currentPosition.Y + 1; i < gameBoard.defenition.YMax; i++ {
			pawn, err := gameBoard.getPawn(position(currentPosition.X, i))
			if err == nil {
				return pawn, nil
			}
		}
		return Pawn{}, errors.New("no next pawn")
	}

	if currentDirection == DOWN {
		for i := currentPosition.Y - 1; i > 0; i-- {
			pawn, err := gameBoard.getPawn(position(currentPosition.X, i))
			if err == nil {
				return pawn, nil
			}
		}
		return Pawn{}, errors.New("no next pawn")
	}

	if !isWithinBoard(gameBoard.Pawns, currentPosition.Y) {
		return Pawn{}, errors.New("invalid current position")
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
			return Pawn{}, errors.New("no next pawn")
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
			return Pawn{}, errors.New("no next pawn")
		}
	}
	return Pawn{}, errors.New("no next pawn")

}

func (gameBoard GameBoard) getFinalDirection(startingPosition Position, startingDirection int) int {
	deflections := gameBoard.GetDeflections(startingPosition, startingDirection)
	deflectionCount := len(deflections)
	return deflections[deflectionCount-1].ToDirection
}

func (gameBoard GameBoard) GetDeflections(startingPosition Position, startingDirection int) []Deflection {
	currentDirection := startingDirection
	currentPosition := startingPosition

	deflections := []Deflection{
		{
			Position:    currentPosition,
			ToDirection: currentDirection,
		},
	}

	// being stuck in an infinite loop is not possible when given valid inputs.
	// an upperbound of len(events) * 2 protects against the possibility of
	// an infinite loop in case unexpected inputs are passed in
	for i := 0; i < len(gameBoard.defenition.Events)*2; i++ {
		pawn, err := gameBoard.getNextPawn(currentPosition, currentDirection)
		if err != nil {
			return deflections
		}
		currentPosition = pawn.Position
		currentDirection = pawn.getDeflectedDirection(currentDirection)
		deflections = append(deflections, Deflection{
			Position:    currentPosition,
			ToDirection: currentDirection,
		})
	}
	return deflections
}
