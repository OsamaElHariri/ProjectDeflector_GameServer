package gamemechanics

import "errors"

const (
	CREATE_PAWN    = "create_pawn"
	FIRE_DEFLECTOR = "fire_deflector"
	SKIP_PAWN      = "skip_pawn"
	END_TURN       = "end_turn"
	MATCH_POINT    = "match_point"
	GAME_WIN       = "game_win"
)

type ProcessedGameBoard struct {
	PlayersInMatchPoint  map[string]bool
	AvailableShuffles    map[string]int
	GameBoard            GameBoard
	ProcessingEventIndex int
	LastDeflections      []Deflection
	VarianceFactory      VarianceFactory
	GameInProgress       bool
	Winner               string
	PawnVariants         map[string][]string
}

type GameEvent interface {
	UpdateGameBoard(gameBoardInProcess ProcessedGameBoard) (ProcessedGameBoard, error)
	Encode() map[string]interface{}
	Decode(anyMap map[string]interface{}) GameEvent
}

func DecodeGameEvent(props map[string]interface{}) (GameEvent, error) {
	if props["name"] == CREATE_PAWN {
		return (CreatePawnEvent{}).Decode(props), nil
	} else if props["name"] == FIRE_DEFLECTOR {
		return (FireDeflectorEvent{}).Decode(props), nil
	} else if props["name"] == SKIP_PAWN {
		return (SkipPawnEvent{}).Decode(props), nil
	} else if props["name"] == END_TURN {
		return (EndTurnEvent{}).Decode(props), nil
	} else if props["name"] == MATCH_POINT {
		return (MatchPointEvent{}).Decode(props), nil
	} else if props["name"] == GAME_WIN {
		return (WinEvent{}).Decode(props), nil
	}

	return CreatePawnEvent{}, errors.New("could not parse game event")
}
