// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

import (
	"strconv"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
)

const (
	CharCornerTopLeft int = iota
	CharCornerTopRight
	CharCornerBottomLeft
	CharCornerBottomRight
	CharHorizontal
	CharVertical
)

var heavyDialog = [...]rune{
	201, // CP437 BOX DRAWINGS DOUBLE DOWN AND RIGHT
	187, // CP437 BOX DRAWINGS DOUBLE DOWN AND LEFT
	200, // CP437 BOX DRAWINGS DOUBLE UP AND RIGHT
	188, // CP437 BOX DRAWINGS DOUBLE UP AND LEFT
	205, // CP437 BOX DRAWINGS DOUBLE HORIZONTAL
	186, // CP437 BOX DRAWINGS DOUBLE VERTICAL
}

var lightDialog = [...]rune{
	218, // CP437 BOX DRAWINGS LIGHT DOWN AND RIGHT
	191, // CP437 BOX DRAWINGS LIGHT DOWN AND LEFT
	192, // CP437 BOX DRAWINGS LIGHT UP AND RIGHT
	217, // CP437 BOX DRAWINGS LIGHT UP AND LEFT
	196, // CP437 BOX DRAWINGS LIGHT HORIZONTAL
	179, // CP437 BOX DRAWINGS LIGHT VERTICAL
}

func (ui *UI) inRange(x, y int) bool {
	return x >= 0 && y >= 0 && x < ui.cols && y < ui.rows
}

func (ui *UI) renderDialog(weight [6]rune, title string, x, y, w, h int, scrollPos float64, a float64) {
	c := &ui.LabelColor
	bg := &ui.BGColor
	if a != 1 {
		c = c.Mul(a)
		bg = bg.Mul(a)
	}

	// Corners first
	if ui.inRange(x, y) {
		t := &ui.textBuffer[x+y*ui.cols]
		t.Color = c
		t.BGColor = bg
		t.Rune = weight[CharCornerTopLeft]
	}
	if ui.inRange(x+w, y) {
		t := &ui.textBuffer[(x+w)+y*ui.cols]
		t.Color = c
		t.BGColor = bg
		t.Rune = weight[CharCornerTopRight]
	}
	if ui.inRange(x, y+h-1) {
		t := &ui.textBuffer[x+(y+h-1)*ui.cols]
		t.Color = c
		t.BGColor = bg
		t.Rune = weight[CharCornerBottomLeft]
	}
	if ui.inRange(x+w, y+h-1) {
		t := &ui.textBuffer[(x+w)+(y+h-1)*ui.cols]
		t.Color = c
		t.BGColor = bg
		t.Rune = weight[CharCornerBottomRight]
	}

	titleStart := x + w/2 - len(title)/2
	titleEnd := titleStart + len(title)

	for i := x + 1; i < x+w; i++ {
		// Top bar first
		if !ui.inRange(i, y) {
			continue
		}
		t := &ui.textBuffer[i+y*ui.cols]
		t.Color = c
		t.BGColor = bg

		switch {
		case i == x+1:
			t.Rune = '['
			t.Shadow = ui.ShadowText
		case i == x+2:
			t.Rune = 254 // CP437 BLACK SQUARE
			t.Shadow = ui.ShadowText
		case i == x+3:
			t.Rune = ']'
			t.Shadow = ui.ShadowText
		case i == titleStart-1 || i == titleEnd:
			t.Rune = 0
		case i >= titleStart && i < titleEnd:
			t.Rune = []rune(title)[i-titleStart]
			t.Shadow = ui.ShadowText
		default:
			t.Rune = weight[CharHorizontal]
		}

		// Bottom bar
		if !ui.inRange(i, y+h-1) {
			continue
		}
		t = &ui.textBuffer[i+(y+h-1)*ui.cols]
		t.Color = c
		t.BGColor = bg
		t.Rune = weight[CharHorizontal]

		// Fill
		for j := y + 1; j < y+h-1; j++ {
			if !ui.inRange(i, j) {
				continue
			}
			t := &ui.textBuffer[i+j*ui.cols]
			t.Color = c
			t.BGColor = bg
			t.Rune = 0
		}
	}

	yScroll := int(scrollPos * float64(h-5))
	hasScroll := scrollPos >= 0
	for i := y + 1; i < y+h-1; i++ {
		// Left bar
		if !ui.inRange(x, i) {
			continue
		}
		t := &ui.textBuffer[x+i*ui.cols]
		t.Color = c
		t.BGColor = bg
		t.Rune = weight[CharVertical]

		// Right bar
		if !ui.inRange(x+w, i) {
			continue
		}
		t = &ui.textBuffer[x+w+i*ui.cols]
		switch {
		case hasScroll && i == y+1:
			t.Color = bg
			t.BGColor = c
			t.Rune = 30 // CP437 BLACK UP-POINTING TRIANGLE
		case hasScroll && i == y+h-2:
			t.Color = bg
			t.BGColor = c
			t.Rune = 31 // CP437 BLACK UP-POINTING TRIANGLE
		case hasScroll && i == y+2+yScroll:
			t.Color = bg
			t.BGColor = c
			t.Rune = 254 // CP437 BLACK SQUARE
		case hasScroll:
			t.Color = bg
			t.BGColor = c
			t.Rune = 177 // CP437 LIGHT SHADE
		default:
			t.Color = c
			t.BGColor = bg
			t.Rune = weight[CharVertical]
		}
	}
}

func (ui *UI) renderBox(widget *Widget, label string, x, y, hStart, hEnd int) {
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
			ui.textBuffer[index].BGColor = &widget.highlight.Render
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

func (ui *UI) measureBox(label string) (int, int) {
	return ui.Padding*2 + len(label) + 1, 2
}

func (ui *UI) renderButton(b *Button, x, y int) {
	ui.renderBox(&b.Widget, b.Label, x, y, -1, -1)
}

func (ui *UI) measureButton(b *Button) (int, int) {
	return ui.measureBox(b.Label)
}

func (ui *UI) renderCheckbox(cb *Checkbox, x, y int) {
	label := cb.Label + " ["
	if cb.Value {
		label += "X]"
	} else {
		label += " ]"
	}
	ui.renderBox(&cb.Widget, label, x, y, ui.Padding+len(cb.Label)+1, ui.Padding+len(label))
}

func (ui *UI) measureCheckbox(cb *Checkbox) (int, int) {
	return ui.measureBox(cb.Label + " [X]")
}

func (ui *UI) measureSlider(s *Slider) (int, int) {
	label := s.Label + " [ " + strconv.Itoa(s.Value) + string(rune(29)) + "]"

	return ui.measureBox(label)
}

func (ui *UI) renderSlider(s *Slider, x, y int) {
	label := s.Label + " [ " + strconv.Itoa(s.Value) + string(rune(29)) + "]"

	ui.renderBox(&s.Widget, label, x, y, ui.Padding+len(s.Label)+1, ui.Padding+len(label))
}

func (ui *UI) renderTooltip(p *Page, widget *Widget) {
	w, h := ui.MeasureString(ui.TextStyle, widget.Tooltip)
	w /= ui.TextStyle.CharWidth
	h /= ui.TextStyle.CharHeight

	x := ui.cols / 2
	y := ui.rows - h - ui.Padding*2 - 1

	ui.renderDialog(lightDialog, "?",
		x-w/2-ui.Padding, y,
		w+ui.Padding*2, h+ui.Padding*2, -1, p.tooltipAlpha.Render)
	xx := x - w/2
	for i := 0; i < len(widget.Tooltip); i++ {
		r := []rune(widget.Tooltip)[i]
		if r == '\n' {
			xx = x - w/2
			y++
			continue
		}
		if !ui.inRange(xx, y+2) {
			continue
		}
		t := &ui.textBuffer[xx+(y+2)*ui.cols]
		t.Color = ui.LabelColor.Mul(p.tooltipAlpha.Render)
		t.BGColor = ui.BGColor.Mul(p.tooltipAlpha.Render)
		t.Rune = r
		t.Shadow = ui.ShadowText
		xx++
	}
}

func (ui *UI) measurePage(page *Page) (mx, my, visibleWidgets int) {
	tooltipHeight := 6
	maxHeight := ui.rows - ui.Padding*2 - tooltipHeight
	if page.IsDialog {
		maxHeight -= 3
	}

	mx, my, visibleWidgets = 0, 0, 0
	for _, v := range page.Widgets {
		var dx, dy int
		switch w := v.(type) {
		case *Button:
			dx, dy = ui.measureButton(w)
		case *Checkbox:
			dx, dy = ui.measureCheckbox(w)
		case *Slider:
			dx, dy = ui.measureSlider(w)
		}
		if dx > mx {
			mx = dx
		}
		my += dy + 1
		visibleWidgets++
		if my >= maxHeight {
			break
		}
	}
	return
}

func (ui *UI) renderPage(page *Page) {
	mx, my, visibleWidgets := ui.measurePage(page)
	page.VisibleWidgets = visibleWidgets
	x := ui.cols / 2
	y := ui.rows/2 - my/2
	if page.IsDialog {
		scrollPos := -1.0
		if page.VisibleWidgets != len(page.Widgets) {
			scrollPos = float64(page.ScrollPos) / float64(len(page.Widgets)-page.VisibleWidgets)
		}
		mx += ui.Padding*2 + 2
		ui.renderDialog(heavyDialog, page.Title,
			x-mx/2, y-ui.Padding,
			mx, my+ui.Padding, scrollPos, 1)
	}

	limit := concepts.Min(len(page.Widgets), page.ScrollPos+visibleWidgets)
	for i := page.ScrollPos; i < limit; i++ {
		selected := i == page.SelectedItem
		switch w := page.Widgets[i].(type) {
		case *Button:
			ui.renderButton(w, x, y)
		case *Checkbox:
			ui.renderCheckbox(w, x, y)
		case *Slider:
			ui.renderSlider(w, x, y)
		}
		y += 3
		w := page.Widgets[i].GetWidget()
		if selected {
			w.highlight.Animation.Lifetime = dynamic.AnimationLifetimeBounce
			w.highlight.Animation.Active = true
		} else {
			w.highlight.Animation.Lifetime = dynamic.AnimationLifetimeBounceOnce
		}
	}
	if page.tooltipCurrent != nil {
		ui.renderTooltip(page, page.tooltipCurrent.GetWidget())
	}
	if page.tooltipQueue.Len() > 0 {
		if page.tooltipAlpha.Percent >= 1 {
			page.tooltipAlpha.Reverse = true
			page.tooltipAlpha.Active = true
		} else if page.tooltipAlpha.Percent <= 0 {
			page.tooltipCurrent = page.tooltipQueue.PopBack()
			page.tooltipAlpha.Reverse = false
			page.tooltipAlpha.Active = true
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
	s.HSpacing = 0
	s.VSpacing = 0
	if s.SpriteSheet == nil || s.CharWidth == 0 || s.CharHeight == 0 {
		return
	}
	img := materials.GetImage(s.SpriteSheet.Universe, s.SpriteSheet.Material)
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
