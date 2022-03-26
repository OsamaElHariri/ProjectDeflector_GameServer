package gamemechanics

import (
	"errors"
	"time"
)

type EndTurnEvent struct {
	name        string
	playerOwner string
	endTime     int64
}

func NewEndTurnEvent(playerOwner string) EndTurnEvent {
	return EndTurnEvent{
		name:        END_TURN,
		playerOwner: playerOwner,
		endTime:     time.Now().UnixMilli(),
	}
}

func (event EndTurnEvent) UpdateGameBoard(gameBoardInProcess ProcessedGameBoard) (ProcessedGameBoard, error) {
	currentPlayer := GetPlayerTurn(gameBoardInProcess.GameBoard)
	expired := gameBoardInProcess.LastTurnEndTime+int64(gameBoardInProcess.GameBoard.defenition.TimePerTurn) < time.Now().UnixMilli()
	if event.playerOwner != currentPlayer && !expired {
		return ProcessedGameBoard{}, errors.New("cannot end the turn of another player")
	}

	gameBoardInProcess.GameBoard.Turn += 1

	nextPlayerTurn := GetPlayerTurn(gameBoardInProcess.GameBoard)
	gameBoardInProcess.AvailableShuffles[nextPlayerTurn] = 1
	if gameBoardInProcess.GameBoard.ScoreBoard[nextPlayerTurn] < gameBoardInProcess.GameBoard.defenition.TargetScore {
		gameBoardInProcess.GameBoard.ScoreBoard[nextPlayerTurn] += 1
	}

	gameBoardInProcess.LastTurnEndTime = event.endTime

	return gameBoardInProcess, nil
}

func (event EndTurnEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"name":        event.name,
		"playerOwner": event.playerOwner,
		"endTime":     event.endTime,
	}
}

func (event EndTurnEvent) Decode(anyMap map[string]interface{}) GameEvent {
	event.name = anyMap["name"].(string)
	event.playerOwner = anyMap["playerOwner"].(string)
	event.endTime = anyMap["endTime"].(int64)

	return event
}
