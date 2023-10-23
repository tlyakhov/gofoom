package core

type CollisionResponse int

//go:generate go run github.com/dmarkham/enumer -type=CollisionResponse -json
const (
	Slide CollisionResponse = iota
	Bounce
	Stop
	Remove
	Callback
)
