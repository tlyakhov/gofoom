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

//go:generate go run github.com/dmarkham/enumer -type=BodyShadow -json
type BodyShadow int

const (
	BodyShadowSphere BodyShadow = iota
	BodyShadowAABB
)

//go:generate go run github.com/dmarkham/enumer -type=ScriptStyle -json
type ScriptStyle int

const (
	ScriptStyleRaw ScriptStyle = iota
	ScriptStyleBoolExpr
	ScriptStyleStatement
)
