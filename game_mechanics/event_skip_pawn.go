package gamemechanics

import "errors"

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
	currentPlayer := GetPlayerTurn(gameBoardInProcess.GameBoard)
	if event.playerOwner != currentPlayer {
		return ProcessedGameBoard{}, errors.New("out of turn action")
	}

	shuffleCount := gameBoardInProcess.AvailableShuffles[event.playerOwner]

	if shuffleCount <= 0 {
		return gameBoardInProcess, errors.New("out of shuffles for turn")
	}

	gameBoardInProcess.AvailableShuffles[event.playerOwner] -= 1
	variants := gameBoardInProcess.PawnVariants[event.playerOwner]

	gameBoardInProcess.PawnVariants[event.playerOwner] = gameBoardInProcess.VarianceFactory.GeneratePawnVariant(getPlayerDigest(gameBoardInProcess.GameBoard.defenition, event.playerOwner), len(variants)+1)
	return gameBoardInProcess, nil
}

func (event SkipPawnEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"name":        event.name,
		"playerOwner": event.playerOwner,
	}
}

func (event SkipPawnEvent) Decode(anyMap map[string]interface{}) GameEvent {
	event.name = anyMap["name"].(string)
	event.playerOwner = anyMap["playerOwner"].(string)

	return event
}
