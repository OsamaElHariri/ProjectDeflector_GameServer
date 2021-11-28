package gamemechanics

const (
	CREATE_PAWN = "create"
	DESTROY     = "destroy"
)

type GameEvent struct {
	name       string
	position   Position
	targetType string
	owner      string
}

func NewGameEvent(x int, y int, targetType string) GameEvent {
	return GameEvent{
		name:       CREATE_PAWN,
		position:   Position{X: x, Y: y},
		targetType: targetType,
		owner:      "anyone",
	}
}
