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

// TODO: Add ability to anchor string left/mid/right, top/mid/bot
func (r *Renderer) DrawString(s *TextStyle, x, y int, text string) {
	if s.Sprite == nil || s.Width == 0 || s.Height == 0 {
		return
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
				//c := concepts.Vector4{float64(col + row), fw, fh}
				a := s.Sprite.SampleAlpha(
					float64(u)*fw,
					float64(v)*fh,
					uint32(s.Width),
					uint32(s.Height),
					col, row)
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
