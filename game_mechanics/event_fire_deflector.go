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

func (event FireDeflectorEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"name": event.name,
	}
}

func (event FireDeflectorEvent) Decode(anyMap map[string]interface{}) GameEvent {
	event.name = anyMap["name"].(string)

	return event
}
