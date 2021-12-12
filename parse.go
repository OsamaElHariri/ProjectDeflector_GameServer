package main

import (
	gamemechanics "projectdeflector/game/game_mechanics"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Pawn struct {
	Position    Position `json:"position"`
	Name        string   `json:"name"`
	Durability  int      `json:"durability"`
	PlayerOwner string   `json:"playerOwner"`
}

type GameBoard struct {
	XMin       int        `json:"xMin"`
	XMax       int        `json:"xMax"`
	Pawns      [][]Pawn   `json:"pawns"`
	Turn       int        `json:"turn"`
	ScoreBoard ScoreBoard `json:"scoreBoard"`
}

type Deflection struct {
	Position    Position          `json:"position"`
	ToDirection int               `json:"toDirection"`
	Events      []DeflectionEvent `json:"events"`
}

type DeflectionEvent struct {
	Name     string   `json:"name"`
	Position Position `json:"position"`
}

type ScoreBoard struct {
	Red  int `json:"red"`
	Blue int `json:"blue"`
}

func parseDeflections(deflections []gamemechanics.Deflection) []Deflection {
	parsedDeflections := make([]Deflection, len(deflections))

	for i, deflection := range deflections {
		parsedDeflections[i] = Deflection{
			Position:    parsePosition(deflection.Position),
			ToDirection: deflection.ToDirection,
			Events:      parseDeflectionEvents(deflection.Events),
		}
	}

	return parsedDeflections
}

func parseDeflectionEvents(deflectionEvents []gamemechanics.DeflectionEvent) []DeflectionEvent {
	parsedDeflectionEvents := make([]DeflectionEvent, len(deflectionEvents))

	for i, deflectionEvent := range deflectionEvents {
		parsedDeflectionEvents[i] = DeflectionEvent{
			Name:     deflectionEvent.Name,
			Position: parsePosition(deflectionEvent.Position),
		}
	}

	return parsedDeflectionEvents
}

func parseGameBoard(gameBoard gamemechanics.GameBoard) GameBoard {
	pawns := make([][]Pawn, len(gameBoard.Pawns))

	for i := 0; i < len(gameBoard.Pawns); i++ {
		for j := 0; j < len(gameBoard.Pawns[i]); j++ {
			pawns[i] = append(pawns[i], parsePawn(*gameBoard.Pawns[i][j]))
		}
	}

	return GameBoard{
		XMin:       gameBoard.XMin,
		XMax:       gameBoard.XMax,
		Turn:       gameBoard.Turn,
		Pawns:      pawns,
		ScoreBoard: parseScoreBoard(gameBoard.ScoreBoard),
	}
}

func parsePawn(pawn gamemechanics.Pawn) Pawn {
	return Pawn{
		Position:    parsePosition(pawn.Position),
		Name:        pawn.Name,
		Durability:  pawn.Durability,
		PlayerOwner: pawn.PlayerOwner,
	}
}

func parsePosition(position gamemechanics.Position) Position {
	return Position{X: position.X, Y: position.Y}
}

func parseScoreBoard(scoreBoard gamemechanics.ScoreBoard) ScoreBoard {
	return ScoreBoard{Red: scoreBoard.Red, Blue: scoreBoard.Blue}
}

func parsePlayerTurn(player int) string {
	if player == gamemechanics.RED_SIDE {
		return "red"
	} else {
		return "blue"
	}
}
