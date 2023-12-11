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

type MaterialScale int

//go:generate go run github.com/dmarkham/enumer -type=MaterialScale -json
const (
	ScaleNone MaterialScale = iota
	ScaleHeight
	ScaleWidth
	ScaleAll
)

//go:generate go run github.com/dmarkham/enumer -type=BodyShadow -json
type BodyShadow int

const (
	BodyShadowSphere BodyShadow = iota
	BodyShadowAABB
)
