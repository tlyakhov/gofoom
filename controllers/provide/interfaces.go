package provide

import "tlyakhov/gofoom/core"

type Interactable interface {
	ActOnMob(e core.AbstractMob)
}

type Passable interface {
	OnEnter(e core.AbstractMob)
	OnExit(e core.AbstractMob)
	Collide(e core.AbstractMob)
	Recalculate()
	UpdatePVS()
}

type Animateable interface {
	Frame()
}

type Hurtable interface {
	Hurt(amount float64)
	HurtTime() float64
}

type Collideable interface {
	Collide() []*core.Segment
}
