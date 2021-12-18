package gamemechanics

const (
	CREATE_PAWN    = "create_pawn"
	FIRE_DEFLECTOR = "fire_deflector"
)

type ProcessedGameBoard struct {
	GameBoard            GameBoard
	ProcessingEventIndex int
	LastDeflections      []Deflection
	VariantFactory       PawnVariantFactory
}

type GameEvent interface {
	UpdateGameBoard(gameBoardInProcess ProcessedGameBoard) (ProcessedGameBoard, error)
	DoesConsumeVariant(playerId string) bool
}
