package provide

type animator interface {
	For(concrete interface{}) Animateable
}
type interactor interface {
	For(concrete interface{}) Interactable
}

type passer interface {
	For(concrete interface{}) Passable
}

type hurter interface {
	For(concrete interface{}) Hurtable
}

var SectorAnimator animator
var EntityAnimator animator
var Interactor interactor
var Passer passer
var Hurter hurter
