package provide

import "github.com/tlyakhov/gofoom/mapping"

type Interactable interface {
	ActOnEntity(e mapping.AbstractEntity)
}

type Passable interface {
	OnEnter(e mapping.AbstractEntity)
	OnExit(e mapping.AbstractEntity)
	Collide(e mapping.AbstractEntity)
}

type Animateable interface {
	Frame(lastFrameTime float64)
}
