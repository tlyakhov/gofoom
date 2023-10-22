package provide

type animator interface {
	For(model interface{}) Animateable
}
type interactor interface {
	For(model interface{}) Interactable
}

type passer interface {
	For(model interface{}) Passable
}

type hurter interface {
	For(model interface{}) Hurtable
}

type collider interface {
	For(model interface{}) Collideable
}

var SectorAnimator animator
var MobAnimator animator
var Interactor interactor
var Passer passer
var Hurter hurter
var Collider collider
