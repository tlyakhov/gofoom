// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

type textElement struct {
	Rune    rune
	Color   *concepts.Vector4
	BGColor *concepts.Vector4
}
type UIConfig struct {
	LabelColor    concepts.Vector4
	SelectedColor concepts.DynamicValue[concepts.Vector4]
	BGColor       concepts.Vector4
	ShadowColor   concepts.Vector4
	TextStyle     *TextStyle
}
type UI struct {
	UIConfig
	*Renderer

	Page *UIPage

	textBuffer []textElement
}

type UIElement struct {
	Label  string
	Action func()
}

type UIPage struct {
	Items        []UIElement
	SelectedItem int
}

func (ui *UI) Initialize() {
	ui.LabelColor = concepts.Vector4{1, 1, 1, 1}
	ui.BGColor = concepts.Vector4{0.3, 0.3, 0.3, 1}
	ui.ShadowColor = concepts.Vector4{0.1, 0.1, 0.1, 1}

	ui.SelectedColor.Attach(ui.DB.Simulation)
	a := ui.SelectedColor.NewAnimation()
	ui.SelectedColor.Animation = a
	a.Duration = 400
	a.Start = ui.BGColor
	a.End = concepts.Vector4{0, 0.431, 1, 1}
	a.TweeningFunc = concepts.EaseOut2

	ui.TextStyle = ui.NewTextStyle()
	sw := ui.ScreenWidth / ui.TextStyle.CharWidth
	sh := ui.ScreenHeight / ui.TextStyle.CharHeight

	if len(ui.textBuffer) != sw*sh {
		ui.textBuffer = make([]textElement, sw*sh)
	}
}

func (ui *UI) Button(label string, x, y int, selected bool) {
	var index int
	bgcolor := &ui.UIConfig.BGColor
	if selected {
		bgcolor = ui.SelectedColor.Render
	}
	sw := ui.ScreenWidth / ui.TextStyle.CharWidth
	sh := ui.ScreenHeight / ui.textStyle.CharHeight
	padding := 2
	extent := concepts.Max(10, len(label)+padding*2)
	x -= extent / 2
	for i := 0; i < extent; i++ {
		if x+i < 0 || x+i >= sw || y < 0 || y >= sh {
			continue
		}
		index = x + i + y*sw
		ui.textBuffer[index].Color = &ui.LabelColor
		ui.textBuffer[index].BGColor = bgcolor
		if i < padding || i >= len(label)+padding {
			ui.textBuffer[index].Rune = ' '
		} else {
			ui.textBuffer[index].Rune = []rune(label)[i-padding]
		}
		if y+1 >= sh {
			continue
		}
		shadow := index + sw + 1
		ui.textBuffer[shadow].Color = &ui.ShadowColor
		ui.textBuffer[shadow].BGColor = nil
		ui.textBuffer[shadow].Rune = 223 // CP437 UPPER HALF BLOCK
	}
	index = x + extent + y*sw
	ui.textBuffer[index].Color = &ui.ShadowColor
	ui.textBuffer[index].BGColor = nil
	ui.textBuffer[index].Rune = 220 // CP437 LOWER HALF BLOCK
}

func (ui *UI) renderPage(page *UIPage) {
	sw := ui.ScreenWidth / ui.TextStyle.CharWidth
	sh := ui.ScreenHeight / ui.TextStyle.CharHeight
	x := sw / 2
	y := sh/2 - len(page.Items)*3/2
	for i, item := range page.Items {
		ui.Button(item.Label, x, y+i*3, i == page.SelectedItem)
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
	}
	ui.renderPage(ui.Page)
	ui.renderTextBuffer()
}

func (ui *UI) renderTextBuffer() {
	s := ui.TextStyle
	s.Shadow = true
	s.HAnchor = -1
	s.VAnchor = -1
	if s.Sprite == nil || s.CharWidth == 0 || s.CharHeight == 0 {
		return
	}
	img := materials.ImageFromDb(s.Sprite.DB, s.Sprite.Image)
	if img == nil {
		return
	}

	c := s.Color
	sw := ui.ScreenWidth / s.CharWidth
	sh := ui.ScreenHeight / s.CharHeight
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			t := ui.textBuffer[x+y*sw]
			if t.Rune == 0 {
				continue
			}
			if t.Color != nil {
				s.Color = *t.Color
			} else {
				s.Color = c
			}
			if t.BGColor != nil && t.BGColor[3] > 0 {
				s.BGColor = *t.BGColor
			} else {
				s.BGColor[3] = 0
			}
			ui.drawChar(s, img, t.Rune, x*s.CharWidth, y*s.CharHeight)
		}
	}
}
