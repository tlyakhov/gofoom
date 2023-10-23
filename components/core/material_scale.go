package core

type MaterialScale int

//go:generate go run github.com/dmarkham/enumer -type=MaterialScale -json
const (
	ScaleNone MaterialScale = iota
	ScaleHeight
	ScaleWidth
	ScaleAll
)
