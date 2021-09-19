package main

import (
	gamemechanics "projectdeflector/game/game_mechanics"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Pawn struct {
	Position      Position `json:"position"`
	Name          string   `json:"name"`
	TurnPlaced    int      `json:"turnPlaced"`
	TurnDestroyed int      `json:"turnDestroyed"`
}

type GameBoard struct {
	XMin  int      `json:"xMin"`
	XMax  int      `json:"xMax"`
	Pawns [][]Pawn `json:"pawns"`
}

type Deflection struct {
	Position    Position `json:"position"`
	ToDirection int      `json:"toDirection"`
}

func parseDeflections(deflections []gamemechanics.Deflection) []Deflection {
	parsedDeflections := make([]Deflection, len(deflections))

	for i, deflection := range deflections {
		parsedDeflections[i] = Deflection{
			Position:    parsePosition(deflection.Position),
			ToDirection: deflection.ToDirection,
		}
	}

	return parsedDeflections
}

func parseGameBoard(gameBoard gamemechanics.GameBoard) GameBoard {
	pawns := make([][]Pawn, len(gameBoard.Pawns))

	for i := 0; i < len(gameBoard.Pawns); i++ {
		for j := 0; j < len(gameBoard.Pawns[i]); j++ {
			pawns[i] = append(pawns[i], parsePawn(gameBoard.Pawns[i][j]))
		}
	}

	return GameBoard{
		XMin:  gameBoard.XMin,
		XMax:  gameBoard.XMax,
		Pawns: pawns,
	}
}

func parsePawn(pawn gamemechanics.Pawn) Pawn {
	return Pawn{
		Position:      parsePosition(pawn.Position),
		Name:          pawn.Name,
		TurnPlaced:    pawn.TurnPlaced,
		TurnDestroyed: pawn.TurnDestroyed,
	}
}

func parsePosition(position gamemechanics.Position) Position {
	return Position{X: position.X, Y: position.Y}
}
