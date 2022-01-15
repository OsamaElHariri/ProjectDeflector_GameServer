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
	variants := gameBoardInProcess.PawnVariants[event.playerOwner]

	gameBoardInProcess.PawnVariants[event.playerOwner] = gameBoardInProcess.VarianceFactory.GeneratePawnVariant(getPlayerDigest(gameBoardInProcess.GameBoard.defenition, event.playerOwner), len(variants)+1)
	return gameBoardInProcess, nil
}
