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

	nextPlayerTurn := GetPlayerTurn(gameBoardInProcess.GameBoard)
	gameBoardInProcess.AvailableShuffles[nextPlayerTurn] = 1
	gameBoardInProcess.GameBoard.ScoreBoard[nextPlayerTurn] += 1

	return gameBoardInProcess, nil
}
