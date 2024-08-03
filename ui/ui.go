// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render"
)

type textElement struct {
	Rune    rune
	Color   *concepts.Vector4
	BGColor *concepts.Vector4
}
type Config struct {
	LabelColor    concepts.Vector4
	SelectedColor concepts.Vector4
	BGColor       concepts.Vector4
	ShadowColor   concepts.Vector4
	TextStyle     *render.TextStyle
}
type UI struct {
	Config
	*render.Renderer

	Page *Page

	textBuffer []textElement
}

func (ui *UI) SetPage(page *Page) {
	if ui.Page != nil {
		for _, item := range ui.Page.Items {
			switch x := item.(type) {
			case *Button:
				x.bgColor.Detach(ui.DB.Simulation)
			}
		}
	}

	ui.Page = page

	for _, item := range page.Items {
		switch x := item.(type) {
		case *Button:
			x.bgColor.Attach(ui.DB.Simulation)
			x.bgColor.SetAll(ui.BGColor)
			a := x.bgColor.NewAnimation()
			x.bgColor.Animation = a
			a.Duration = 400
			a.Start = ui.BGColor
			a.End = ui.SelectedColor
			a.TweeningFunc = concepts.EaseOut2
			a.Active = false
			a.Coordinates = concepts.AnimationCoordinatesAbsolute
		}
	}
}

func (ui *UI) Initialize() {
	ui.LabelColor = concepts.Vector4{1, 1, 1, 1}
	ui.BGColor = concepts.Vector4{0.3, 0.3, 0.3, 1}
	ui.ShadowColor = concepts.Vector4{0.1, 0.1, 0.1, 1}
	ui.SelectedColor = concepts.Vector4{0, 0.431, 1, 1}

	ui.TextStyle = ui.NewTextStyle()
	sw := ui.ScreenWidth / ui.TextStyle.CharWidth
	sh := ui.ScreenHeight / ui.TextStyle.CharHeight

	if len(ui.textBuffer) != sw*sh {
		ui.textBuffer = make([]textElement, sw*sh)
	}
}

func (ui *UI) Button(button *Button, x, y int) {
	var index int
	sw := ui.ScreenWidth / ui.TextStyle.CharWidth
	sh := ui.ScreenHeight / ui.TextStyle.CharHeight
	padding := 2
	extent := concepts.Max(10, len(button.Label)+padding*2)
	x -= extent / 2
	for i := 0; i < extent; i++ {
		if x+i < 0 || x+i >= sw || y < 0 || y >= sh {
			continue
		}
		index = x + i + y*sw
		ui.textBuffer[index].Color = &ui.LabelColor
		ui.textBuffer[index].BGColor = button.bgColor.Render
		if i < padding || i >= len(button.Label)+padding {
			ui.textBuffer[index].Rune = ' '
		} else {
			ui.textBuffer[index].Rune = []rune(button.Label)[i-padding]
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

func (ui *UI) renderPage(page *Page) {
	sw := ui.ScreenWidth / ui.TextStyle.CharWidth
	sh := ui.ScreenHeight / ui.TextStyle.CharHeight
	x := sw / 2
	y := sh/2 - len(page.Items)*3/2
	for i, v := range page.Items {
		selected := i == page.SelectedItem
		switch item := v.(type) {
		case *Button:
			ui.Button(item, x, y+i*3)
			if selected {
				item.bgColor.Animation.Lifetime = concepts.AnimationLifetimeBounce
				item.bgColor.Animation.Active = true
			} else {
				item.bgColor.Animation.Lifetime = concepts.AnimationLifetimeBounceOnce
			}
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
			ui.DrawChar(s, img, t.Rune, x*s.CharWidth, y*s.CharHeight)
		}
	}
}
