// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

type TextStyle struct {
	Sprite                *materials.Sprite
	CharWidth, CharHeight int
	HSpacing, VSpacing    int
	VAnchor, HAnchor      int
	Color                 concepts.Vector4
	BGColor               concepts.Vector4
	Shadow                bool
	ClipX, ClipY          int
	ClipW, ClipH          int

	sample concepts.Vector4
	shadow concepts.Vector4
}

func (r *Renderer) NewTextStyle() *TextStyle {
	return &TextStyle{
		Sprite:     r.DefaultFont(),
		CharWidth:  8,
		CharHeight: 8,
		HSpacing:   0,
		VSpacing:   0,
		Color:      concepts.Vector4{1, 1, 1, 1},
		Shadow:     false,
		ClipW:      r.ScreenWidth,
		ClipH:      r.ScreenHeight,
	}
}

// TODO: This should be more configurable
func (r *Renderer) DefaultFont() *materials.Sprite {
	// TODO: Avoid searching every time
	return materials.SpriteFromDb(r.DB, r.DB.GetEntityByName("HUD Font"))
}

func (r *Renderer) DrawChar(s *TextStyle, img *materials.Image, c rune, dx, dy int) {
	fw := 1.0 / float64(s.CharWidth)
	fh := 1.0 / float64(s.CharHeight)
	index := uint32(c)
	col := index % s.Sprite.Cols
	row := index / s.Sprite.Cols
	// Background first
	if s.BGColor[3] > 0 {
		bgx := dx
		bgy := dy
		for v := 0; v < s.CharHeight; v++ {
			for u := 0; u < s.CharWidth; u++ {
				// Clip to screen
				if bgx < s.ClipX || bgx >= s.ClipX+s.ClipW ||
					bgy < s.ClipY || bgy >= s.ClipY+s.ClipH {
					bgx++
					continue
				}
				r.ApplySample(&s.BGColor, bgx+bgy*r.ScreenWidth, -1)
				bgx++
			}
			bgx -= s.CharWidth
			bgy++
		}
	}
	if c == 0 {
		return
	}
	// Foreground
	for v := 0; v < s.CharHeight; v++ {
		for u := 0; u < s.CharWidth; u++ {
			// Clip to screen
			if dx < s.ClipX || dx >= s.ClipX+s.ClipW ||
				dy < s.ClipY || dy >= s.ClipY+s.ClipH {
				dx++
				continue
			}
			screenIndex := dx + dy*r.ScreenWidth
			ur, vr := s.Sprite.TransformUV(float64(u)*fw, float64(v)*fh, col, row)
			a := img.SampleAlpha(ur, vr,
				uint32(s.CharWidth)*s.Sprite.Cols,
				uint32(s.CharHeight)*s.Sprite.Rows)
			a *= s.Color[3]
			if s.Shadow && dx < s.ClipX+s.ClipW-1 && dy < s.ClipY+s.ClipH-1 {
				s.shadow[3] = a * 0.5
				r.ApplySample(&s.shadow, screenIndex+1+r.ScreenWidth, -2)
			}
			s.sample[3] = a
			s.sample[2] = s.Color[2] * a
			s.sample[1] = s.Color[1] * a
			s.sample[0] = s.Color[0] * a
			r.ApplySample(&s.sample, screenIndex, -3)
			dx++
		}
		dx -= s.CharWidth
		dy++
	}
}

func (r *Renderer) Print(s *TextStyle, x, y int, text string) {
	if s.Sprite == nil || s.CharWidth == 0 || s.CharHeight == 0 {
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
	for _, c := range text {
		if c == '\n' {
			dx = x
			dy += s.CharHeight + s.VSpacing
			continue
		}

		r.DrawChar(s, img, c, dx, dy)
		dx += s.CharWidth + s.HSpacing
	}
}

func (r *Renderer) MeasureString(s *TextStyle, text string) (w int, h int) {
	if s.CharWidth == 0 || s.CharHeight == 0 || len(text) == 0 {
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
			dy += s.CharHeight
		} else {
			// For the 2nd character+ in a line, add a space between letters
			dx += s.HSpacing
		}
		dx += s.CharWidth
	}
	if dx > w {
		w = dx
	}
	h = dy
	return
}
