package gamemechanics

type WinEvent struct {
	name        string
	playerOwner string
}

func NewWinEvent(playerOwner string) WinEvent {
	return WinEvent{
		name:        GAME_WIN,
		playerOwner: playerOwner,
	}
}

func (event WinEvent) UpdateGameBoard(gameBoardInProcess ProcessedGameBoard) (ProcessedGameBoard, error) {
	gameBoardInProcess.GameInProgress = false
	gameBoardInProcess.Winner = event.playerOwner

	return gameBoardInProcess, nil
}
