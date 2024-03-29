package gamemechanics

type Position struct {
	X int
	Y int
}

func NewPosition(x int, y int) Position {
	return position(x, y)
}

func position(x int, y int) Position {
	return Position{
		X: x,
		Y: y,
	}
}

func (pos Position) equals(pos2 Position) bool {
	return pos.X == pos2.X && pos.Y == pos2.Y
}

func (pos Position) toMap() map[string]interface{} {
	return map[string]interface{}{
		"x": pos.X,
		"y": pos.Y,
	}
}
