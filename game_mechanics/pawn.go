package main

const (
	SLASH     = "slash"
	BACKSLASH = "backslash"
)

type Pawn struct {
	position      Position
	name          string
	turnPlaced    int
	turnDestroyed int
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

	if pawn.name == BACKSLASH {
		return backslashDeflection[currentDirection]
	}

	if pawn.name == SLASH {
		return slashDeflection[currentDirection]
	}
	return currentDirection
}
