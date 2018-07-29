package mapping

type MaterialBehavior int

//go:generate stringer -type=MaterialBehavior
const (
	ScaleNone MaterialBehavior = iota
	ScaleWidth
	ScaleHeight
	ScaleAll
)
