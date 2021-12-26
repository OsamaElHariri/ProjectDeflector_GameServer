package gamemechanics

const (
	CREATE_PAWN    = "create_pawn"
	FIRE_DEFLECTOR = "fire_deflector"
	SKIP_PAWN      = "skip_pawn"
	END_TURN       = "end_turn"
)

type ProcessedGameBoard struct {
	GameBoard            GameBoard
	ProcessingEventIndex int
	LastDeflections      []Deflection
	VarianceFactory      VarianceFactory
}

type GameEvent interface {
	UpdateGameBoard(gameBoardInProcess ProcessedGameBoard) (ProcessedGameBoard, error)
	DoesConsumeVariant(playerId string) bool
}
