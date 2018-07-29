package mapping

type Player struct {
	AliveEntity

	Height    float64
	Standing  bool
	Crouching bool
	Inventory []Entity
}
