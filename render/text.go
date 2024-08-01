// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

type TextStyle struct {
	Sprite             *materials.Sprite
	Width, Height      int
	HSpacing, VSpacing int
	VAnchor, HAnchor   int
	Color              concepts.Vector4
	Shadow             bool
	ClipX, ClipY       int
	ClipW, ClipH       int
}

func (r *Renderer) NewTextStyle() *TextStyle {
	return &TextStyle{
		Sprite:   r.DefaultFont(),
		Width:    8,
		Height:   8,
		HSpacing: 0,
		VSpacing: 0,
		Color:    concepts.Vector4{1, 1, 1, 1},
		Shadow:   false,
		ClipW:    r.ScreenWidth,
		ClipH:    r.ScreenHeight,
	}
}

// TODO: This should be more configurable
func (r *Renderer) DefaultFont() *materials.Sprite {
	// TODO: Avoid searching every time
	return materials.SpriteFromDb(r.DB, r.DB.GetEntityByName("HUD Font"))
}

func (r *Renderer) DrawString(s *TextStyle, x, y int, text string) {
	if s.Sprite == nil || s.Width == 0 || s.Height == 0 {
		return
	}
	img := materials.ImageFromDb(s.Sprite.DB, s.Sprite.Image)
	if img == nil {
		return
	}
	var mx, my int

	if s.HAnchor >= 0 || s.VAnchor >= 0 {
		mx, my = r.MeasureString(s, text)
		if s.HAnchor == 0 {
			x += -mx / 2
		} else {
			x += -mx
		}
		if s.VAnchor == 0 {
			y += -my / 2
		} else {
			y += -my
		}
	}
	dx := x
	dy := y
	fw := 1.0 / float64(s.Width)
	fh := 1.0 / float64(s.Height)
	for _, c := range text {
		if c == '\n' {
			dx = x
			dy += s.Height + s.VSpacing
			continue
		}
		index := uint32(c)
		col := index % s.Sprite.Cols
		row := index / s.Sprite.Cols
		for v := 0; v < s.Height; v++ {
			for u := 0; u < s.Width; u++ {
				// Clip to screen
				if dx < s.ClipX || dx >= s.ClipX+s.ClipW ||
					dy < s.ClipY || dy >= s.ClipY+s.ClipH {
					dx++
					continue
				}
				screenIndex := dx + dy*r.ScreenWidth
				ur, vr := s.Sprite.TransformUV(float64(u)*fw, float64(v)*fh, col, row)
				a := img.SampleAlpha(ur, vr,
					uint32(s.Width)*s.Sprite.Cols,
					uint32(s.Height)*s.Sprite.Rows)
				a *= s.Color[3]
				if s.Shadow && dx < s.ClipX+s.ClipW-1 && dy < s.ClipY+s.ClipH-1 {
					r.ApplySample(&concepts.Vector4{0, 0, 0, a}, screenIndex+1+r.ScreenWidth, -1)
				}
				r.ApplySample(&concepts.Vector4{s.Color[0] * a, s.Color[1] * a, s.Color[2] * a, a}, screenIndex, -2)
				dx++
			}
			dx -= s.Width
			dy++
		}
		dx += s.Width + s.HSpacing
		dy -= s.Height
	}
}

func (r *Renderer) MeasureString(s *TextStyle, text string) (w int, h int) {
	if s.Width == 0 || s.Height == 0 || len(text) == 0 {
		return
	}
	w = 0
	h = 0
	dx := 0
	dy := 0
	for _, c := range text {
		if c == '\n' {
			if dx > w {
				w = dx
			}
			dx = 0
			dy += s.VSpacing
			continue
		}
		if dx == 0 {
			// For the first character in a line, include the line height
			dy += s.Height
		} else {
			// For the 2nd character+ in a line, add a space between letters
			dx += s.HSpacing
		}
		dx += s.Width
	}
	if dx > w {
		w = dx
	}
	h = dy
	return
}
