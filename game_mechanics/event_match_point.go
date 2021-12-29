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

func (event MatchPointEvent) DoesConsumeVariant(playerId string) bool {
	return false
}
