package gamemechanics

import (
	"testing"
)

type PredictableVarianceFactory struct {
	variants map[string][]string
}

func (factory PredictableVarianceFactory) GeneratePawnVariant(str string, turns int) []string {
	return factory.variants[str][0:turns]
}

func (factory PredictableVarianceFactory) GenerateDeflectionSource(gameBoard GameBoard, turn int) DirectedPosition {
	return DirectedPosition{
		Position:  position(gameBoard.defenition.XMax/2, -1),
		Direction: UP,
	}
}

func TestNewGameBoard(t *testing.T) {

	processedGameBoard, err := NewGameBoard(GameBoardDefenition{
		PlayerIds: []string{"red", "blue"},
		YMax:      2,
		XMax:      4,
		Events: []GameEvent{
			NewCreatePawnEvent(position(2, 1), "red"),
			NewCreatePawnEvent(position(3, 2), "blue"),
		},
	})
	if err != nil || len(processedGameBoard.GameBoard.Pawns) < 2 || len(processedGameBoard.GameBoard.Pawns[1]) < 4 {
		t.Errorf("Failed to create board")
	}
	gameBoard := processedGameBoard.GameBoard
	pawn, err := gameBoard.getPawn(position(2, 1))
	if err != nil || pawn.Position.X != 2 || pawn.Position.Y != 1 {
		t.Errorf("Failed to get pawn")
	}

	pawn, err = gameBoard.getPawn(position(3, 2))
	if err != nil || pawn.Position.X != 3 || pawn.Position.Y != 2 {
		t.Errorf("Failed to get pawn")
	}

	pawn, err = gameBoard.getPawn(position(0, 1))
	if pawn != nil || err == nil {
		t.Errorf("Got pawn when there is no pawn")
	}

	pawn, err = gameBoard.getPawn(position(500, 500))
	if pawn != nil || err == nil {
		t.Errorf("Got pawn when there is no pawn")
	}
}

func TestPawnTraversal(t *testing.T) {
	processedGameBoard, err := NewGameBoard(GameBoardDefenition{
		PlayerIds: []string{"red", "blue"},
		YMax:      2,
		XMax:      4,
		Events: []GameEvent{
			NewCreatePawnEvent(position(2, 2), "red"),
			NewCreatePawnEvent(position(2, 1), "blue"),
			NewCreatePawnEvent(position(2, 0), "red"),
			NewCreatePawnEvent(position(0, 1), "blue"),
			NewCreatePawnEvent(position(4, 1), "red"),
		},
	})

	if err != nil {
		t.Errorf("Failed to created game board")
	}

	gameBoard := processedGameBoard.GameBoard
	pawn, err := gameBoard.getNextPawn(position(2, 1), UP)
	if err != nil || !pawn.Position.equals(position(2, 2)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.Position.X, pawn.Position.Y)
	}

	pawn, err = gameBoard.getNextPawn(position(2, 1), DOWN)
	if err != nil || !pawn.Position.equals(position(2, 0)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.Position.X, pawn.Position.Y)
	}

	pawn, err = gameBoard.getNextPawn(position(2, 1), LEFT)
	if err != nil || !pawn.Position.equals(position(0, 1)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.Position.X, pawn.Position.Y)
	}

	pawn, err = gameBoard.getNextPawn(position(2, 1), RIGHT)
	if err != nil || !pawn.Position.equals(position(4, 1)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.Position.X, pawn.Position.Y)
	}

	pawn, err = gameBoard.getNextPawn(position(2, -1), UP)
	if err != nil || !pawn.Position.equals(position(2, 0)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.Position.X, pawn.Position.Y)
	}

	_, err = gameBoard.getNextPawn(position(2, 0), DOWN)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	_, err = gameBoard.getNextPawn(position(4, 1), DOWN)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	_, err = gameBoard.getNextPawn(position(2, 0), LEFT)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	_, err = gameBoard.getNextPawn(position(2, 2), UP)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	_, err = gameBoard.getNextPawn(position(0, 1), UP)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	_, err = gameBoard.getNextPawn(position(4, 1), RIGHT)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	_, err = gameBoard.getNextPawn(position(2, 2), RIGHT)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}
}

func TestGetFinalDirection(t *testing.T) {
	processedGameBoard, err := newGameBoard(GameBoardDefenition{
		PlayerIds: []string{"red", "blue"},
		YMax:      2,
		XMax:      4,
		Events: []GameEvent{
			NewCreatePawnEvent(position(2, 0), "red"),
			NewCreatePawnEvent(position(1, 0), "blue"),
			NewCreatePawnEvent(position(1, 2), "red"),
			NewCreatePawnEvent(position(3, 2), "blue"),
			NewCreatePawnEvent(position(3, 0), "red"),
			NewCreatePawnEvent(position(2, 1), "blue"),
			NewFireDeflectorEvent(),
		},
	}, PredictableVarianceFactory{
		variants: map[string][]string{
			"0red":  {BACKSLASH, SLASH, SLASH, SLASH, SLASH},
			"0blue": {BACKSLASH, BACKSLASH, SLASH, SLASH, SLASH},
		},
	})

	if err != nil {
		t.Errorf("Failed to created game board")
	}

	deflections := processedGameBoard.LastDeflections
	if deflections[len(deflections)-1].ToDirection != RIGHT {
		t.Errorf("Wrong final direction %d", deflections[len(deflections)-1].ToDirection)
	}
}
