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
	SelectedColor concepts.Vector4
	BGColor       concepts.Vector4
	ShadowColor   concepts.Vector4
	TextStyle     *TextStyle
}
type UI struct {
	UIConfig
	*Renderer

	textBuffer []textElement
}

func (ui *UI) Initialize() {
	ui.LabelColor = concepts.Vector4{1, 1, 1, 1}
	ui.BGColor = concepts.Vector4{0.2, 0.2, 0.2, 1}
	ui.ShadowColor = concepts.Vector4{0, 0, 0, 1}
	ui.SelectedColor = concepts.Vector4{0.5, 0.5, 1, 1}
	ui.TextStyle = ui.NewTextStyle()
	sw := ui.ScreenWidth / ui.TextStyle.CharWidth
	sh := ui.ScreenHeight / ui.TextStyle.CharHeight

	if len(ui.textBuffer) != sw*sh {
		ui.textBuffer = make([]textElement, sw*sh)
	}
}

func (ui *UI) Button(label string, x, y int, selected bool) {
	var index int
	color := &ui.LabelColor
	if selected {
		color = &ui.SelectedColor
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
		ui.textBuffer[index].Color = color
		ui.textBuffer[index].BGColor = &ui.UIConfig.BGColor
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

func (ui *UI) NewFrame() {
	for i := 0; i < len(ui.textBuffer); i++ {
		ui.textBuffer[i].Rune = 0
		ui.textBuffer[i].BGColor = nil
		ui.textBuffer[i].Color = nil
	}
}

func (ui *UI) RenderTextBuffer() {
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
