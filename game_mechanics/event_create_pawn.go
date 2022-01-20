package gamemechanics

type CreatePawnEvent struct {
	name        string
	position    Position
	playerOwner string
}

func NewCreatePawnEvent(pos Position, playerOwner string) CreatePawnEvent {
	return CreatePawnEvent{
		name:        CREATE_PAWN,
		position:    pos,
		playerOwner: playerOwner,
	}
}

func (event CreatePawnEvent) UpdateGameBoard(gameBoardInProcess ProcessedGameBoard) (ProcessedGameBoard, error) {
	variants := gameBoardInProcess.PawnVariants[event.playerOwner]
	variant := variants[len(variants)-1]

	newPawn := Pawn{
		Position:    event.position,
		Name:        variant,
		TurnPlaced:  gameBoardInProcess.GameBoard.Turn,
		Durability:  5,
		PlayerOwner: event.playerOwner,
	}
	gameBoardInProcess.PawnVariants[event.playerOwner] = gameBoardInProcess.VarianceFactory.GeneratePawnVariant(getPlayerDigest(gameBoardInProcess.GameBoard.defenition, event.playerOwner), len(variants)+1)
	updatedPawns, err := addPawn(gameBoardInProcess.GameBoard.Pawns, newPawn)
	if err != nil {
		return gameBoardInProcess, err
	}
	gameBoardInProcess.GameBoard.Pawns = updatedPawns

	gameBoardInProcess.GameBoard.ScoreBoard[event.playerOwner] -= 1

	return gameBoardInProcess, nil
}
