package main

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
	yMax   int
	events []GameEvent
}

type GameBoard struct {
	defenition GameBoardDefenition
	xMin       int
	xMax       int
	pawns      [][]Pawn
}

func newGameBoard(defenition GameBoardDefenition) (GameBoard, error) {
	gameBoard := GameBoard{
		defenition: defenition,
		pawns:      make([][]Pawn, defenition.yMax),
	}
	for _, event := range defenition.events {
		gameBoard = applyEvent(gameBoard, event)
	}

	return gameBoard, nil
}

func (gameBoard GameBoard) getPawn(position Position) (Pawn, error) {
	index, err := getPawnIndex(gameBoard.pawns, position)
	if err != nil {
		return Pawn{}, err
	}
	return gameBoard.pawns[position.y][index], nil
}

func getPawnIndex(pawns [][]Pawn, position Position) (int, error) {
	if !isWithinBoard(pawns, position.y) {
		return 0, errors.New("out of bounds")
	}

	for i, pawn := range pawns[position.y] {
		if pawn.position.x == position.x {
			return i, nil
		}
	}

	return 0, errors.New("empty position")
}

func isWithinBoard(pawns [][]Pawn, yCoord int) bool {
	return yCoord >= 0 && yCoord < len(pawns)
}

func applyEvent(gameBoard GameBoard, event GameEvent) GameBoard {
	if event.name == CREATE {
		updatedPawns, err := addPawn(gameBoard.pawns, event)
		if err == nil {
			gameBoard.pawns = updatedPawns

		}
	}
	return gameBoard
}

func addPawn(pawns [][]Pawn, event GameEvent) ([][]Pawn, error) {
	if !isWithinBoard(pawns, event.position.y) {
		return pawns, errors.New("invalid pawn position")
	}

	newPawn := Pawn{
		position:   event.position,
		name:       event.targetType,
		turnPlaced: 0,
	}

	index, err := getPawnIndex(pawns, event.position)
	if err == nil {
		pawns[event.position.y][index] = newPawn
	} else {
		pawns[event.position.y] = append(pawns[event.position.y], newPawn)
	}

	return pawns, nil
}

func (gameBoard GameBoard) getNextPawn(currentPosition Position, currentDirection int) (Pawn, error) {

	if currentDirection == UP {
		for i := currentPosition.y + 1; i < gameBoard.defenition.yMax; i++ {
			pawn, err := gameBoard.getPawn(position(currentPosition.x, i))
			if err == nil {
				return pawn, nil
			}
		}
		return Pawn{}, errors.New("no next pawn")
	}

	if currentDirection == DOWN {
		for i := currentPosition.y - 1; i > 0; i-- {
			pawn, err := gameBoard.getPawn(position(currentPosition.x, i))
			if err == nil {
				return pawn, nil
			}
		}
		return Pawn{}, errors.New("no next pawn")
	}

	if !isWithinBoard(gameBoard.pawns, currentPosition.y) {
		return Pawn{}, errors.New("invalid current position")
	}

	if currentDirection == LEFT {
		indexNearest := -1
		for index, pawn := range gameBoard.pawns[currentPosition.y] {
			if pawn.position.x < currentPosition.x {
				if indexNearest < 0 || (indexNearest >= 0 && pawn.position.x > gameBoard.pawns[currentPosition.y][indexNearest].position.x) {
					indexNearest = index
				}
			}
		}
		if indexNearest >= 0 {
			return gameBoard.pawns[currentPosition.y][indexNearest], nil
		} else {
			return Pawn{}, errors.New("no next pawn")
		}
	}

	if currentDirection == RIGHT {
		indexNearest := -1
		for index, pawn := range gameBoard.pawns[currentPosition.y] {
			if pawn.position.x > currentPosition.x {
				if indexNearest < 0 || (indexNearest >= 0 && pawn.position.x < gameBoard.pawns[currentPosition.y][indexNearest].position.x) {
					indexNearest = index
				}
			}
		}
		if indexNearest >= 0 {
			return gameBoard.pawns[currentPosition.y][indexNearest], nil
		} else {
			return Pawn{}, errors.New("no next pawn")
		}
	}
	return Pawn{}, errors.New("no next pawn")

}

func (gameBoard GameBoard) getFinalDirection(startingPosition Position, startingDirection int) int {
	currentDirection := startingDirection
	currentPosition := startingPosition

	// being stuck in an infinite loop is not possible when given valid inputs.
	// an upperbound of len(events) * 2 protects against the possibility of
	// an infinite loop in case unexpected inputs are passed in
	for i := 0; i < len(gameBoard.defenition.events)*2; i++ {
		pawn, err := gameBoard.getNextPawn(currentPosition, currentDirection)
		if err != nil {
			return currentDirection
		}
		currentPosition = pawn.position
		currentDirection = pawn.getDeflectedDirection(currentDirection)
	}
	return currentDirection
}
