// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"fmt"
	"time"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
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
		r.BitBlt(img, 10, r.ScreenHeight-42, 32, 32)
	}
}

func (r *Renderer) DebugInfo() {
	//defer concepts.ExecutionDuration(concepts.ExecutionTrack("DebugInfo"))

	playerAlive := behaviors.AliveFromDb(r.DB, r.PlayerBody.Entity)
	// player := bodies.PlayerFromDb(&gameMap.Player)

	ts := r.NewTextStyle()
	ts.Color[3] = 0.5
	ts.Shadow = true

	for x := 0; x < constants.RenderBlocks; x++ {
		c := r.Columns[x]
		for _, b := range c.BodiesSeen {
			top := &concepts.Vector3{}
			top[0] = b.Pos.Render[0]
			top[1] = b.Pos.Render[1]
			top[2] = b.Pos.Render[2] + b.Size.Render[1]*0.5
			scr := r.WorldToScreen(top)
			if scr == nil {
				continue
			}
			text := fmt.Sprintf("%v", b.String())
			r.DrawString(ts, int(scr[0]), int(scr[1])-16, text)
		}
	}

	r.DrawString(ts, 4, 4, fmt.Sprintf("FPS: %.1f, Light cache: %v", r.DB.Simulation.FPS, r.SectorLastRendered.Size()))
	r.DrawString(ts, 4, 14, fmt.Sprintf("Health: %.1f", playerAlive.Health))
	hits := r.ICacheHits.Load()
	misses := r.ICacheMisses.Load()
	r.DrawString(ts, 4, 24, fmt.Sprintf("ICache hit percentage: %.1f, %v, %v", float64(hits)*100.0/float64(hits+misses), hits, misses))
	if r.PlayerBody.SectorEntity != 0 {
		entity := r.PlayerBody.SectorEntity
		s := 0
		//		core.SectorFromDb(ref).Lightmap.Range(func(k uint64, v concepts.Vector4) bool { s++; return true })
		r.DrawString(ts, 4, 34, fmt.Sprintf("Sector: %v, LM:%v", entity.String(r.DB), s))
		r.DrawString(ts, 4, 44, fmt.Sprintf("f: %v, v: %v, p: %v\n", r.PlayerBody.Force.StringHuman(), r.PlayerBody.Vel.Render.StringHuman(), r.PlayerBody.Pos.Render.StringHuman()))
	}

	for i := 0; i < 20; i++ {
		if i >= r.DebugNotices.Length() {
			break
		}
		msg := r.DebugNotices.Items[i].(string)
		if t, ok := r.DebugNotices.SetWithTimes.Load(msg); ok {
			r.DrawString(ts, 4, 54+i*10, msg)
			age := time.Now().UnixMilli() - t.(int64)
			if age > 10000 {
				r.DebugNotices.PopAtIndex(i)
			}
		}
	}

}
