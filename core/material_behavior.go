package core

type MaterialBehavior int

//go:generate go run github.com/dmarkham/enumer -type=MaterialBehavior -json
const (
	ScaleNone MaterialBehavior = iota
	ScaleHeight
	ScaleWidth
	ScaleAll
)
