package ui

type Button struct {
	Widget

	Clicked func(b *Button)
}
