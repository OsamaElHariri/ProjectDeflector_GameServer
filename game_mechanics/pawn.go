package gamemechanics

const (
	SLASH     = "slash"
	BACKSLASH = "backslash"
)

type Pawn struct {
	Position    Position
	Name        string
	TurnPlaced  int
	Durability  int
	PlayerOwner string
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

func (pawn Pawn) toMap() map[string]interface{} {
	return map[string]interface{}{
		"position":    pawn.Position.toMap(),
		"name":        pawn.Name,
		"turnPlaced":  pawn.TurnPlaced,
		"durability":  pawn.Durability,
		"playerOwner": pawn.PlayerOwner,
	}
}
