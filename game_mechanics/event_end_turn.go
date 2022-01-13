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

	gameBoardInProcess.GameBoard.ScoreBoard[event.playerOwner] += 1
	return gameBoardInProcess, nil
}
