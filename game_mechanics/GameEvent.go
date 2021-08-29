package main

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
