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
	turnsPlayed := getTurnsPlayed(gameBoardInProcess.GameBoard.defenition.Events[0:gameBoardInProcess.ProcessingEventIndex], event.playerOwner)

	variant := gameBoardInProcess.VariantFactory.Generate(gameBoardInProcess.GameBoard.GetPlayerDigest(event.playerOwner), turnsPlayed+1)

	newPawn := Pawn{
		Position:    event.position,
		Name:        variant[len(variant)-1],
		TurnPlaced:  gameBoardInProcess.GameBoard.Turn,
		Durability:  3,
		PlayerOwner: event.playerOwner,
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
