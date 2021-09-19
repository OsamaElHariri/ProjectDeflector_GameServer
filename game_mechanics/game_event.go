package gamemechanics

const (
	CREATE  = "create"
	DESTROY = "destroy"
)

type GameEvent struct {
	name       string
	position   Position
	targetType string
	owner      string
}

func NewGameEvent(x int, y int, targetType string) GameEvent {
	return GameEvent{
		name:       CREATE,
		position:   Position{X: x, Y: y},
		targetType: targetType,
		owner:      "anyone",
	}
}
