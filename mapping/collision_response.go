package mapping

type CollisionResponse int

//go:generate enumer -type=CollisionResponse -json
const (
	Slide CollisionResponse = iota
	Bounce
	Stop
	Remove
	Callback
)
