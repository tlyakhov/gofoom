// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import "tlyakhov/gofoom/components/materials"

// TODO: This should be more configurable
func (r *Renderer) DefaultFont() *materials.Sprite {
	// TODO: Avoid searching every time
	return materials.SpriteFromDb(r.DB, r.DB.GetEntityByName("HUD Font"))
}

// TODO: Wrap all parameters in a struct
// TODO: Add ability to anchor string left/mid/right, top/mid/bot
func (r *Renderer) DrawString(sprite *materials.Sprite, s string, x, y, w, h int) {
	if sprite == nil {
		return
	}
	for _, c := range s {
		index := uint32(c)
		col := index % sprite.Cols
		row := index / sprite.Cols
		for v := 0; v < h; v++ {
			for u := 0; u < w; u++ {
				if x < 0 || x >= r.ScreenWidth || y < 0 || y >= r.ScreenHeight {
					x++
					continue
				}
				screenIndex := x + y*r.ScreenWidth
				c := sprite.Sample(float64(u)/float64(w), float64(v)/float64(h), uint32(w), uint32(h), col, row)
				r.ApplySample(&c, screenIndex, -1)
				x++
			}
			x -= w
			y++
		}
		x += w
		y -= h
	}
}
