package provide

import "tlyakhov/gofoom/core"

type Interactable interface {
	ActOnEntity(e core.AbstractEntity)
}

type Passable interface {
	OnEnter(e core.AbstractEntity)
	OnExit(e core.AbstractEntity)
	Collide(e core.AbstractEntity)
	Recalculate()
	UpdatePVS()
}

type Animateable interface {
	Frame(lastFrameTime float64)
}

type Hurtable interface {
	Hurt(amount float64)
	HurtTime() float64
}

type Collideable interface {
	Collide()
}
