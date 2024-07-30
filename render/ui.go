// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"fmt"
	"time"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

func (r *Renderer) RenderHud() {
	if r.Player == nil {
		return
	}
	for _, item := range r.Player.Inventory {
		img := materials.ImageFromDb(r.DB, item.Image)
		if img == nil {
			return
		}
		r.BitBlt(img, 10, r.ScreenHeight-42, 64, 64)
	}
}

func (r *Renderer) DebugInfo() {
	playerAlive := behaviors.AliveFromDb(r.DB, r.PlayerBody.Entity)
	// player := bodies.PlayerFromDb(&gameMap.Player)

	font := r.DefaultFont()
	screen := r.WorldToScreen(&concepts.Vector3{-320, 30, -32})
	if screen != nil {
		r.DrawString(font, "world:-320,30,-32", int(screen[0]), int(screen[1]), 8, 8)
	}
	r.DrawString(font, fmt.Sprintf("FPS: %.1f, Light cache: %v", r.DB.Simulation.FPS, r.SectorLastRendered.Size()), 4, 4, 8, 8)
	r.DrawString(font, fmt.Sprintf("Health: %.1f", playerAlive.Health), 4, 14, 8, 8)
	hits := r.ICacheHits.Load()
	misses := r.ICacheMisses.Load()
	r.DrawString(font, fmt.Sprintf("ICache hit percentage: %.1f, %v, %v", float64(hits)*100.0/float64(hits+misses), hits, misses), 4, 24, 8, 8)
	if r.PlayerBody.SectorEntity != 0 {
		entity := r.PlayerBody.SectorEntity
		s := 0
		//		core.SectorFromDb(ref).Lightmap.Range(func(k uint64, v concepts.Vector4) bool { s++; return true })
		r.DrawString(font, fmt.Sprintf("Sector: %v, LM:%v", entity.String(r.DB), s), 4, 34, 8, 8)
		r.DrawString(font, fmt.Sprintf("f: %v, v: %v, p: %v\n", r.PlayerBody.Force.StringHuman(), r.PlayerBody.Vel.Render.StringHuman(), r.PlayerBody.Pos.Render.StringHuman()), 4, 44, 8, 8)
	}

	for i := 0; i < 20; i++ {
		if i >= r.DebugNotices.Length() {
			break
		}
		msg := r.DebugNotices.Items[i].(string)
		if t, ok := r.DebugNotices.SetWithTimes.Load(msg); ok {
			r.DrawString(font, msg, 4, 54+i*10, 8, 8)
			age := time.Now().UnixMilli() - t.(int64)
			if age > 10000 {
				r.DebugNotices.PopAtIndex(i)
			}
		}
	}

}
