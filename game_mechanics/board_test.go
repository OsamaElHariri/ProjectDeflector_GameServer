package main

import "testing"

func TestNewGameBoard(t *testing.T) {

	gameBoard, err := newGameBoard(GameBoardDefenition{
		yMax: 5,
		events: []GameEvent{
			{
				name:     CREATE,
				position: position(1, 1),
			},
			{
				name:     CREATE,
				position: position(1, 1),
			},
			{
				name:     CREATE,
				position: position(500, 2),
			},
		},
	})
	if err != nil || len(gameBoard.pawns) < 5 || len(gameBoard.pawns[1]) < 1 {
		t.Errorf("Failed to create board")
	}

	pawn, err := gameBoard.getPawn(position(1, 1))
	if err != nil || pawn.position.x != 1 || pawn.position.y != 1 {
		t.Errorf("Failed to get pawn")
	}

	pawn, err = gameBoard.getPawn(position(500, 2))
	if err != nil || pawn.position.x != 500 || pawn.position.y != 2 {
		t.Errorf("Failed to get pawn")
	}

	pawn, err = gameBoard.getPawn(position(0, 1))
	if err == nil {
		t.Errorf("Got pawn when there is no pawn")
	}

	pawn, err = gameBoard.getPawn(position(1, 0))
	if err == nil {
		t.Errorf("Got pawn when there is no pawn")
	}

}

func TestPawnTraversal(t *testing.T) {
	gameBoard, err := newGameBoard(GameBoardDefenition{
		yMax: 5,
		events: []GameEvent{
			{
				name:       CREATE,
				targetType: BACKSLASH,
				position:   position(1, 1),
			},
			{
				name:       CREATE,
				targetType: SLASH,
				position:   position(1, 4),
			},
			{
				name:       CREATE,
				targetType: BACKSLASH,
				position:   position(7, 4),
			},
			{
				name:       CREATE,
				targetType: BACKSLASH,
				position:   position(3, 4),
			},
			{
				name:       CREATE,
				targetType: BACKSLASH,
				position:   position(-1, 4),
			},
		},
	})

	if err != nil {
		t.Errorf("Failed to created game board")
	}

	pawn, err := gameBoard.getNextPawn(position(1, 1), UP)
	if err != nil || !pawn.position.equals(position(1, 4)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.position.x, pawn.position.y)
	}

	pawn, err = gameBoard.getNextPawn(position(7, 7), DOWN)
	if err != nil || !pawn.position.equals(position(7, 4)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.position.x, pawn.position.y)
	}

	pawn, err = gameBoard.getNextPawn(position(1, 4), LEFT)
	if err != nil || !pawn.position.equals(position(-1, 4)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.position.x, pawn.position.y)
	}

	pawn, err = gameBoard.getNextPawn(position(1, 4), RIGHT)
	if err != nil || !pawn.position.equals(position(3, 4)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.position.x, pawn.position.y)
	}

	pawn, err = gameBoard.getNextPawn(position(1, 1), DOWN)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	pawn, err = gameBoard.getNextPawn(position(-1, 4), LEFT)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	pawn, err = gameBoard.getNextPawn(position(7, 4), RIGHT)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	pawn, err = gameBoard.getNextPawn(position(2, 2), UP)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	pawn, err = gameBoard.getNextPawn(position(2, 2), DOWN)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	pawn, err = gameBoard.getNextPawn(position(2, 2), LEFT)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	pawn, err = gameBoard.getNextPawn(position(2, 2), RIGHT)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

	pawn, err = gameBoard.getNextPawn(position(2, 1000), RIGHT)
	if err == nil {
		t.Errorf("Got next pawn when there is no pawn")
	}

}

func TestGetFinalDirection(t *testing.T) {
	gameBoard, err := newGameBoard(GameBoardDefenition{
		yMax: 5,
		events: []GameEvent{
			{
				name:       CREATE,
				targetType: BACKSLASH,
				position:   position(1, 1),
			},
			{
				name:       CREATE,
				targetType: SLASH,
				position:   position(-2, 1),
			},
		},
	})

	if err != nil {
		t.Errorf("Failed to created game board")
	}

	finalDirection := gameBoard.getFinalDirection(position(1, 0), UP)
	if finalDirection != DOWN {
		t.Errorf("Wrong simple final direction %d", finalDirection)
	}

	gameBoard, err = newGameBoard(GameBoardDefenition{
		yMax: 5,
		events: []GameEvent{
			{
				name:       CREATE,
				targetType: SLASH,
				position:   position(0, 1),
			},
			{
				name:       CREATE,
				targetType: SLASH,
				position:   position(1, 1),
			},
			{
				name:       CREATE,
				targetType: BACKSLASH,
				position:   position(1, 2),
			},
			{
				name:       CREATE,
				targetType: SLASH,
				position:   position(0, 2),
			},
		},
	})

	if err != nil {
		t.Errorf("Failed to created game board")
	}

	finalDirection = gameBoard.getFinalDirection(position(0, 0), UP)
	if finalDirection != LEFT {
		t.Errorf("Wrong final direction %d", finalDirection)
	}

	finalDirection = gameBoard.getFinalDirection(position(-1, 1), RIGHT)
	if finalDirection != DOWN {
		t.Errorf("Wrong final direction %d", finalDirection)
	}

}
