package gamemechanics

type FireDeflectorEvent struct {
	name string
}

func NewFireDeflectorEvent() FireDeflectorEvent {
	return FireDeflectorEvent{
		name: FIRE_DEFLECTOR,
	}
}

func (event FireDeflectorEvent) UpdateGameBoard(gameBoardInProcess ProcessedGameBoard) (ProcessedGameBoard, error) {
	gameBoardInProcess.GameBoard.Turn += 1

	gameBoard, deflections := ProcessDeflection(gameBoardInProcess.GameBoard)
	gameBoardInProcess.GameBoard = gameBoard
	gameBoardInProcess.LastDeflections = deflections

	return gameBoardInProcess, nil
}

func (event FireDeflectorEvent) DoesConsumeVariant(playerId string) bool {
	return false
}
