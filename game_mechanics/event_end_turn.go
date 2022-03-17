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

func (event EndTurnEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"name":        event.name,
		"playerOwner": event.playerOwner,
	}
}

func (event EndTurnEvent) Decode(anyMap map[string]interface{}) GameEvent {
	event.name = anyMap["name"].(string)
	event.playerOwner = anyMap["playerOwner"].(string)

	return event
}
