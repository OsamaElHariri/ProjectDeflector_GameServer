package gamemechanics

import "testing"

func TestNewGameBoard(t *testing.T) {

	gameBoard, err := NewGameBoard(GameBoardDefenition{
		YMax: 5,
		Events: []GameEvent{
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
	if err != nil || len(gameBoard.Pawns) < 5 || len(gameBoard.Pawns[1]) < 1 {
		t.Errorf("Failed to create board")
	}

	pawn, err := gameBoard.getPawn(position(1, 1))
	if err != nil || pawn.Position.X != 1 || pawn.Position.Y != 1 {
		t.Errorf("Failed to get pawn")
	}

	pawn, err = gameBoard.getPawn(position(500, 2))
	if err != nil || pawn.Position.X != 500 || pawn.Position.Y != 2 {
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
	gameBoard, err := NewGameBoard(GameBoardDefenition{
		YMax: 5,
		Events: []GameEvent{
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
	if err != nil || !pawn.Position.equals(position(1, 4)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.Position.X, pawn.Position.Y)
	}

	pawn, err = gameBoard.getNextPawn(position(7, 7), DOWN)
	if err != nil || !pawn.Position.equals(position(7, 4)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.Position.X, pawn.Position.Y)
	}

	pawn, err = gameBoard.getNextPawn(position(1, 4), LEFT)
	if err != nil || !pawn.Position.equals(position(-1, 4)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.Position.X, pawn.Position.Y)
	}

	pawn, err = gameBoard.getNextPawn(position(1, 4), RIGHT)
	if err != nil || !pawn.Position.equals(position(3, 4)) {
		t.Errorf("Failed to get next pawn, got (%d, %d)", pawn.Position.X, pawn.Position.Y)
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
	gameBoard, err := NewGameBoard(GameBoardDefenition{
		YMax: 5,
		Events: []GameEvent{
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

	gameBoard, err = NewGameBoard(GameBoardDefenition{
		YMax: 5,
		Events: []GameEvent{
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