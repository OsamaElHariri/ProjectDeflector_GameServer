package gamemechanics

type SkipPawnEvent struct {
	name        string
	playerOwner string
}

func NewSkipPawnEvent(playerOwner string) SkipPawnEvent {
	return SkipPawnEvent{
		name:        SKIP_PAWN,
		playerOwner: playerOwner,
	}
}

func (event SkipPawnEvent) UpdateGameBoard(gameBoardInProcess ProcessedGameBoard) (ProcessedGameBoard, error) {
	return gameBoardInProcess, nil
}

func (event SkipPawnEvent) DoesConsumeVariant(playerId string) bool {
	return event.playerOwner == playerId
}
