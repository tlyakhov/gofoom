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

var SectorAnimator animator
var EntityAnimator animator
var Interactor interactor
var Passer passer
