package gamemechanics

const (
	CREATE_PAWN    = "create_pawn"
	FIRE_DEFLECTOR = "fire_deflector"
)

const (
	DESTROY_PAWM = "destroy_pawn"
)

type GameEvent struct {
	name       string
	position   Position
	targetType string
	owner      string
}

func NewGameEvent(name string, x int, y int, targetType string) GameEvent {
	return GameEvent{
		name:       name,
		position:   Position{X: x, Y: y},
		targetType: targetType,
		owner:      "anyone",
	}
}
