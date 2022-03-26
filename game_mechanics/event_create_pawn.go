package gamemechanics

import "errors"

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
	currentPlayer := GetPlayerTurn(gameBoardInProcess.GameBoard)
	if event.playerOwner != currentPlayer {
		return ProcessedGameBoard{}, errors.New("out of turn action")
	}

	if gameBoardInProcess.GameBoard.ScoreBoard[event.playerOwner] <= 0 {
		return ProcessedGameBoard{}, errors.New("out of score")
	}

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

func (event CreatePawnEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"name":        event.name,
		"position_x":  event.position.X,
		"position_y":  event.position.Y,
		"playerOwner": event.playerOwner,
	}
}

func (event CreatePawnEvent) Decode(anyMap map[string]interface{}) GameEvent {
	event.name = anyMap["name"].(string)
	event.playerOwner = anyMap["playerOwner"].(string)
	event.position = position(int(anyMap["position_x"].(int32)), int(anyMap["position_y"].(int32)))

	return event
}
