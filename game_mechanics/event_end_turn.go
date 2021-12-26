package gamemechanics

type EndTurnEvent struct {
	name        string
	playerOwner string
}

func NewEndTurnEvent(playerOwner string) EndTurnEvent {
	return EndTurnEvent{
		name:        END_TURN,
		playerOwner: playerOwner,
	}
}

func (event EndTurnEvent) UpdateGameBoard(gameBoardInProcess ProcessedGameBoard) (ProcessedGameBoard, error) {
	gameBoardInProcess.GameBoard.Turn += 1

	if event.playerOwner == "red" {
		gameBoardInProcess.GameBoard.ScoreBoard.Red += 1
	} else {
		gameBoardInProcess.GameBoard.ScoreBoard.Blue += 1
	}

	return gameBoardInProcess, nil
}

func (event EndTurnEvent) DoesConsumeVariant(playerId string) bool {
	return false
}
