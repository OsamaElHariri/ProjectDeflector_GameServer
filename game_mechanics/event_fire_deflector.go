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
	deflectionSource := gameBoardInProcess.VarianceFactory.GenerateDeflectionSource(gameBoardInProcess.GameBoard, gameBoardInProcess.GameBoard.Turn)
	gameBoard, deflections := ProcessDeflection(gameBoardInProcess.GameBoard, deflectionSource)
	gameBoardInProcess.GameBoard = gameBoard
	gameBoardInProcess.LastDeflections = deflections

	return gameBoardInProcess, nil
}

func (event FireDeflectorEvent) DoesConsumeVariant(playerId string) bool {
	return false
}
