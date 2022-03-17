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

func (event WinEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"name":        event.name,
		"playerOwner": event.playerOwner,
	}
}

func (event WinEvent) Decode(anyMap map[string]interface{}) GameEvent {
	event.name = anyMap["name"].(string)
	event.playerOwner = anyMap["playerOwner"].(string)

	return event
}
