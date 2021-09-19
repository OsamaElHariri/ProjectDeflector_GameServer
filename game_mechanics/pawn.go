package gamemechanics

const (
	SLASH     = "slash"
	BACKSLASH = "backslash"
)

type Pawn struct {
	Position      Position
	Name          string
	TurnPlaced    int
	TurnDestroyed int
}

func (pawn Pawn) getDeflectedDirection(currentDirection int) int {

	backslashDeflection := map[int]int{
		UP:    LEFT,
		LEFT:  UP,
		DOWN:  RIGHT,
		RIGHT: DOWN,
	}

	slashDeflection := map[int]int{
		UP:    RIGHT,
		RIGHT: UP,
		DOWN:  LEFT,
		LEFT:  DOWN,
	}

	if pawn.Name == BACKSLASH {
		return backslashDeflection[currentDirection]
	}

	if pawn.Name == SLASH {
		return slashDeflection[currentDirection]
	}
	return currentDirection
}
