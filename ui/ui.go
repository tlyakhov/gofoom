// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

import (
	"strconv"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render"
)

type Config struct {
	LabelColor    concepts.Vector4
	SelectedColor concepts.Vector4
	BGColor       concepts.Vector4
	WidgetColor   concepts.Vector4
	ShadowColor   concepts.Vector4
	TextStyle     *render.TextStyle
	Padding       int
	ShadowText    bool
}
type UI struct {
	Config
	*render.Renderer

	Page *Page

	cols, rows int
	textBuffer []textElement
}

func (ui *UI) SetPage(page *Page) {
	if ui.Page != nil {
		for _, item := range ui.Page.Widgets {
			w := item.GetWidget()
			w.highlight.Detach(ui.DB.Simulation)
		}
	}

	ui.Page = page

	for _, item := range page.Widgets {
		w := item.GetWidget()
		w.highlight.Attach(ui.DB.Simulation)
		w.highlight.SetAll(ui.WidgetColor)
		a := w.highlight.NewAnimation()
		w.highlight.Animation = a
		a.Duration = 500
		a.Start = w.highlight.Original
		a.End = ui.SelectedColor
		a.TweeningFunc = concepts.EaseOut2
		a.Reverse = true
		a.Active = false
		a.Coordinates = concepts.AnimationCoordinatesAbsolute
	}
}

func (ui *UI) Initialize() {
	ui.Padding = 2
	ui.LabelColor = concepts.Vector4{1, 1, 1, 1}
	ui.BGColor = concepts.Vector4{0.3, 0.3, 0.3, 1}
	ui.WidgetColor = concepts.Vector4{0.5, 0.5, 0.5, 1}
	ui.ShadowColor = concepts.Vector4{0.1, 0.1, 0.1, 1}
	ui.SelectedColor = concepts.Vector4{0, 0.431, 1, 1}
	ui.ShadowText = true

	ui.TextStyle = ui.NewTextStyle()
	ui.cols = ui.ScreenWidth / ui.TextStyle.CharWidth
	ui.rows = ui.ScreenHeight / ui.TextStyle.CharHeight

	if len(ui.textBuffer) != ui.cols*ui.rows {
		ui.textBuffer = make([]textElement, ui.cols*ui.rows)
	}
}
func (ui *UI) inRange(x, y int) bool {
	return x >= 0 && y >= 0 && x < ui.cols && y < ui.rows
}

func (ui *UI) dialog(title string, x, y, w, h int) {
	// Corners first
	if ui.inRange(x, y) {
		t := &ui.textBuffer[x+y*ui.cols]
		t.Color = &ui.LabelColor
		t.BGColor = &ui.BGColor
		t.Rune = 201 // CP437 BOX DRAWINGS DOUBLE DOWN AND RIGHT
	}
	if ui.inRange(x+w, y) {
		t := &ui.textBuffer[(x+w)+y*ui.cols]
		t.Color = &ui.LabelColor
		t.BGColor = &ui.BGColor
		t.Rune = 187 // CP437 BOX DRAWINGS DOUBLE DOWN AND LEFT
	}
	if ui.inRange(x, y+h) {
		t := &ui.textBuffer[x+(y+h)*ui.cols]
		t.Color = &ui.LabelColor
		t.BGColor = &ui.BGColor
		t.Rune = 200 // CP437 BOX DRAWINGS DOUBLE UP AND RIGHT
	}
	if ui.inRange(x+w, y+h) {
		t := &ui.textBuffer[(x+w)+(y+h)*ui.cols]
		t.Color = &ui.LabelColor
		t.BGColor = &ui.BGColor
		t.Rune = 188 // CP437 BOX DRAWINGS DOUBLE UP AND LEFT
	}

	titleStart := x + w/2 - len(title)/2
	titleEnd := titleStart + len(title)

	for i := x + 1; i < x+w; i++ {
		// Top bar first
		if !ui.inRange(i, y) {
			continue
		}
		t := &ui.textBuffer[i+y*ui.cols]
		t.Color = &ui.LabelColor
		t.BGColor = &ui.BGColor

		switch {
		case i == x+2:
			t.Rune = '['
			t.Shadow = ui.ShadowText
		case i == x+3:
			t.Rune = 254 // CP437 BLACK SQUARE
			t.Shadow = ui.ShadowText
		case i == x+4:
			t.Rune = ']'
			t.Shadow = ui.ShadowText
		case i == titleStart-1 || i == titleEnd:
			t.Rune = 0
		case i >= titleStart && i < titleEnd:
			t.Rune = []rune(title)[i-titleStart]
			t.Shadow = ui.ShadowText
		default:
			t.Rune = 205 // CP437 BOX DRAWINGS DOUBLE HORIZONTAL
		}

		// Bottom bar
		if !ui.inRange(i, y+h) {
			continue
		}
		t = &ui.textBuffer[i+(y+h)*ui.cols]
		t.Color = &ui.LabelColor
		t.BGColor = &ui.BGColor
		t.Rune = 205 // CP437 BOX DRAWINGS DOUBLE HORIZONTAL

		// Fill
		for j := y + 1; j < y+h; j++ {
			if !ui.inRange(i, j) {
				continue
			}
			t := &ui.textBuffer[i+j*ui.cols]
			t.Color = &ui.LabelColor
			t.BGColor = &ui.BGColor
			t.Rune = 0
		}
	}

	for i := y + 1; i < y+h; i++ {
		// Left bar
		if !ui.inRange(x, i) {
			continue
		}
		t := &ui.textBuffer[x+i*ui.cols]
		t.Color = &ui.LabelColor
		t.BGColor = &ui.BGColor
		t.Rune = 186 // CP437 BOX DRAWINGS DOUBLE VERTICAL

		// Right bar
		if !ui.inRange(x+w, i) {
			continue
		}
		t = &ui.textBuffer[x+w+i*ui.cols]
		t.Color = &ui.LabelColor
		t.BGColor = &ui.BGColor
		t.Rune = 186 // CP437 BOX DRAWINGS DOUBLE VERTICAL
	}
}

func (ui *UI) box(widget *Widget, label string, x, y, hStart, hEnd int) {
	var index int
	extent := ui.Padding*2 + len(label)
	hStart = concepts.Max(0, hStart)
	if hEnd < 0 || hEnd > ui.Padding*2+len(label) {
		hEnd = ui.Padding*2 + len(label)
	}
	x -= extent / 2
	for i := 0; i < extent; i++ {
		if !ui.inRange(x+i, y) {
			continue
		}
		index = x + i + y*ui.cols
		ui.textBuffer[index].Color = &ui.LabelColor
		if i >= hStart && i < hEnd {
			ui.textBuffer[index].BGColor = widget.highlight.Render
		} else {
			ui.textBuffer[index].BGColor = &ui.WidgetColor
		}

		if i < ui.Padding || i >= extent-ui.Padding {
			ui.textBuffer[index].Rune = 0
		} else {
			ui.textBuffer[index].Rune = []rune(label)[i-ui.Padding]
			ui.textBuffer[index].Shadow = ui.ShadowText
		}

		if ui.inRange(x+i, y+1) {
			shadow := index + ui.cols + 1
			ui.textBuffer[shadow].Color = &ui.ShadowColor
			//ui.textBuffer[shadow].BGColor = nil
			ui.textBuffer[shadow].Rune = 223 // CP437 UPPER HALF BLOCK
		}
	}
	if ui.inRange(x+extent, y) {
		index = x + extent + y*ui.cols
		ui.textBuffer[index].Color = &ui.ShadowColor
		//	ui.textBuffer[index].BGColor = nil
		ui.textBuffer[index].Rune = 220 // CP437 LOWER HALF BLOCK
	}
}

func (ui *UI) Button(button *Button, x, y int) {
	ui.box(&button.Widget, button.Label, x, y, -1, -1)
}

func (ui *UI) Checkbox(cb *Checkbox, x, y int) {
	label := cb.Label + " ["
	if cb.IsChecked {
		label += "X]"
	} else {
		label += " ]"
	}
	ui.box(&cb.Widget, label, x, y, ui.Padding+len(cb.Label)+1, ui.Padding+len(label))
}

func (ui *UI) Slider(cb *Slider, x, y int) {
	label := cb.Label + " [ " + strconv.Itoa(cb.Value) + string(rune(29)) + "]"

	ui.box(&cb.Widget, label, x, y, ui.Padding+len(cb.Label)+1, ui.Padding+len(label))
}

func (ui *UI) renderPage(page *Page) {
	x := ui.cols / 2
	y := ui.rows/2 - len(page.Widgets)*3/2
	ui.dialog("Menu", x-12, y-ui.Padding, 24, len(page.Widgets)*3+ui.Padding)
	for i, v := range page.Widgets {
		selected := i == page.SelectedItem
		switch item := v.(type) {
		case *Button:
			ui.Button(item, x, y+i*3)
		case *Checkbox:
			ui.Checkbox(item, x, y+i*3)
		case *Slider:
			ui.Slider(item, x, y+i*3)
		}
		w := v.GetWidget()
		if selected {
			w.highlight.Animation.Lifetime = concepts.AnimationLifetimeLoop
			w.highlight.Animation.Active = true
		} else {
			w.highlight.Animation.Lifetime = concepts.AnimationLifetimeOnce
		}
	}
}

func (ui *UI) Render() {
	if ui.Page == nil {
		return
	}
	for i := 0; i < len(ui.textBuffer); i++ {
		ui.textBuffer[i].Rune = 0
		ui.textBuffer[i].BGColor = nil
		ui.textBuffer[i].Color = nil
		ui.textBuffer[i].Shadow = false
	}
	ui.renderPage(ui.Page)
	ui.renderTextBuffer()
}

func (ui *UI) renderTextBuffer() {
	s := ui.TextStyle
	s.HAnchor = -1
	s.VAnchor = -1
	if s.Sprite == nil || s.CharWidth == 0 || s.CharHeight == 0 {
		return
	}
	img := materials.ImageFromDb(s.Sprite.DB, s.Sprite.Image)
	if img == nil {
		return
	}

	for y := ui.rows - 1; y >= 0; y-- {
		for x := ui.cols - 1; x >= 0; x-- {
			t := ui.textBuffer[x+y*ui.cols]
			if t.Color != nil {
				s.Color = *t.Color
			} else {
				s.Color[3] = 0
			}
			if t.BGColor != nil && t.BGColor[3] > 0 {
				s.BGColor = *t.BGColor
			} else {
				s.BGColor[3] = 0
			}
			s.Shadow = t.Shadow
			ui.DrawChar(s, img, t.Rune, x*s.CharWidth, y*s.CharHeight)
		}
	}
}
