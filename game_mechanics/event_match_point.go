package gamemechanics

type MatchPointEvent struct {
	name        string
	playerOwner string
}

func NewMatchPointEvent(playerOwner string) MatchPointEvent {
	return MatchPointEvent{
		name:        MATCH_POINT,
		playerOwner: playerOwner,
	}
}

func (event MatchPointEvent) UpdateGameBoard(gameBoardInProcess ProcessedGameBoard) (ProcessedGameBoard, error) {
	gameBoardInProcess.PlayersInMatchPoint[event.playerOwner] = true

	return gameBoardInProcess, nil
}

func (event MatchPointEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"name":         event.name,
		"player_owner": event.playerOwner,
	}
}

func (event MatchPointEvent) Decode(anyMap map[string]interface{}) GameEvent {
	event.name = anyMap["name"].(string)
	event.playerOwner = anyMap["player_owner"].(string)

	return event
}
