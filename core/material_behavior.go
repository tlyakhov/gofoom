package core

type MaterialBehavior int

//go:generate enumer -type=MaterialBehavior -json
const (
	ScaleNone MaterialBehavior = iota
	ScaleHeight
	ScaleWidth
	ScaleAll
)
