package gamemechanics

type CreatePawnEvent struct {
	name        string
	position    Position
	targetType  string
	playerOwner string
}

func NewCreatePawnEvent(pos Position, targetType string, playerOwner string) CreatePawnEvent {
	return CreatePawnEvent{
		name:        CREATE_PAWN,
		position:    pos,
		targetType:  targetType,
		playerOwner: playerOwner,
	}
}

func (event CreatePawnEvent) UpdateGameBoard(gameBoardInProcess ProcessedGameBoard) (ProcessedGameBoard, error) {
	newPawn := Pawn{
		Position:   event.position,
		Name:       event.targetType,
		TurnPlaced: gameBoardInProcess.GameBoard.Turn,
		Durability: 3,
	}
	updatedPawns, err := addPawn(gameBoardInProcess.GameBoard.Pawns, newPawn)
	if err != nil {
		return gameBoardInProcess, err
	}
	gameBoardInProcess.GameBoard.Pawns = updatedPawns

	return gameBoardInProcess, nil
}

func (event CreatePawnEvent) DoesConsumeVariant(playerId string) bool {
	return event.playerOwner == playerId
}
