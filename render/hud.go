// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"fmt"
	"strconv"
	"time"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

func (r *Renderer) RenderHud() {
	if r.Player == nil {
		return
	}
	ts := r.textStyle
	ts.Color[0] = 1
	ts.Color[1] = 1
	ts.Color[2] = 1
	ts.Color[3] = 0.5
	ts.Shadow = true
	ts.HAnchor = 0
	ts.VAnchor = 0

	if r.Player.SelectedTarget != 0 {
		b := core.GetBody(r.ECS, r.Player.SelectedTarget)
		pt := behaviors.GetPlayerTargetable(r.ECS, r.Player.SelectedTarget)
		if pt != nil && len(pt.Message) > 0 {
			top := &concepts.Vector3{}
			top[0] = b.Pos.Render[0]
			top[1] = b.Pos.Render[1]
			top[2] = b.Pos.Render[2] + b.Size.Render[1]*0.5
			scr := r.WorldToScreen(top)
			if scr != nil {
				r.Print(ts, int(scr[0]), int(scr[1])-16, pt.ApplyMessage(r.Player.SelectedTarget))
			}
		}
	}

	for i, slot := range r.Player.Inventory {
		r.BitBlt(slot.Image, i*40+10, r.ScreenHeight-42, 32, 32)
		r.Print(ts, i*40+16+10, r.ScreenHeight-50, strconv.Itoa(slot.Count.Now))
		if slot == r.Player.CurrentWeapon {
			r.BitBlt(slot.Image, r.ScreenWidth/2-64, r.ScreenHeight-128, 128, 128)
		}
	}
}

func (r *Renderer) DebugInfo() {
	//defer concepts.ExecutionDuration(concepts.ExecutionTrack("DebugInfo"))

	playerAlive := behaviors.GetAlive(r.ECS, r.Player.Entity)
	playerMobile := core.GetMobile(r.ECS, r.Player.Entity)

	if playerAlive == nil || playerMobile == nil {
		return
	}

	ts := r.textStyle
	ts.Color[0] = 1
	ts.Color[1] = 1
	ts.Color[2] = 1
	ts.Color[3] = 0.5
	ts.Shadow = true
	ts.HAnchor = 0
	ts.VAnchor = 0
	/*for x := 0; x < r.Blocks; x++ {
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
			if alive := behaviors.GetAlive(r.ECS, b.Entity); alive != nil {
				text = fmt.Sprintf("%v", alive.Health)
			}
			r.Print(ts, int(scr[0]), int(scr[1])-16, text)
		}
	}*/

	ts.HAnchor = -1
	ts.VAnchor = -1

	r.Print(ts, 4, 4, fmt.Sprintf("FPS: %.1f, Total Entities: %v", r.ECS.Simulation.FPS, r.ECS.Entities.Count()))
	r.Print(ts, 4, 14, fmt.Sprintf("Health: %.1f", playerAlive.Health))
	hits := r.ICacheHits.Load()
	misses := r.ICacheMisses.Load()
	r.Print(ts, 4, 24, fmt.Sprintf("ICache hit percentage: %.1f, %v, %v", float64(hits)*100.0/float64(hits+misses), hits, misses))
	if r.PlayerBody.SectorEntity != 0 {
		entity := r.PlayerBody.SectorEntity
		sector := core.GetSector(r.ECS, entity)
		s := sector.Lightmap.Size()
		r.Print(ts, 4, 34, fmt.Sprintf("Sector: %v, LM:%v, Colliders: %v", entity.Format(r.ECS), s, len(sector.Colliders)))
		r.Print(ts, 4, 44, fmt.Sprintf("f: %v, v: %v, p: %v\n", playerMobile.Force.StringHuman(), playerMobile.Vel.Render.StringHuman(), r.PlayerBody.Pos.Render.StringHuman()))
	}

	for i := 0; i < 20; i++ {
		if i >= r.Player.Notices.Length() {
			break
		}
		msg := r.Player.Notices.Items[i]
		if t, ok := r.Player.Notices.SetWithTimes.Load(msg); ok {
			r.Print(ts, 4, 54+i*10, msg)
			age := time.Now().UnixMilli() - t.(int64)
			if age > 10000 {
				r.Player.Notices.PopAtIndex(i)
			}
		}
	}

}
