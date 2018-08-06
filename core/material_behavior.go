package core

type MaterialBehavior int

//go:generate enumer -type=MaterialBehavior -json
const (
	ScaleNone MaterialBehavior = iota
	ScaleWidth
	ScaleHeight
	ScaleAll
)
