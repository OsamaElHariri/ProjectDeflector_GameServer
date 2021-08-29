package main

type Position struct {
	x int
	y int
}

func position(x int, y int) Position {
	return Position{
		x: x,
		y: y,
	}
}

func (pos Position) equals(pos2 Position) bool {
	return pos.x == pos2.x && pos.y == pos2.y
}
