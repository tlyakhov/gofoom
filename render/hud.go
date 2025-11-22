// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

func (r *Renderer) renderWeapon(slot *inventory.Slot) {
	wc := inventory.GetWeaponClass(slot.Entity)
	w := inventory.GetWeapon(slot.Entity)
	if wc != nil && w != nil {
		if w.State == inventory.WeaponFiring {
			r.BitBlt(wc.FlashMaterial, r.ScreenWidth/2-64, r.ScreenHeight-160, 128, 128, concepts.BlendScreen)
		}
		r.BitBlt(wc.Params[w.State].Material, r.ScreenWidth/2-64, r.ScreenHeight-128, 128, 128, concepts.BlendNormal)
	}
}

func (r *Renderer) renderSelectedTarget() {
	ts := r.textStyle

	pt := behaviors.GetPlayerTargetable(r.Player.SelectedTarget)
	if pt == nil || len(pt.Message) == 0 {
		return
	}

	scr := r.WorldToScreen(pt.Pos(r.Player.SelectedTarget))
	if scr == nil {
		return
	}

	params := &PlayerMessageParams{
		TargetableEntity: r.Player.SelectedTarget,
		PlayerTargetable: pt,
		Player:           r.Player,
		Carrier:          r.Carrier,
	}
	msg := strings.TrimSpace(ApplyPlayerMessage(pt, params))
	r.Print(ts, int(scr[0]), int(scr[1])-16, msg)
}

func (r *Renderer) renderHUD() {
	if r.Player == nil || r.Carrier == nil {
		return
	}

	// All visible bodies
	visited := ecs.EntityTable{}
	for i := range r.Blocks {
		block := &r.Blocks[i]
		for b := range block.Bodies {
			if visited.Contains(b.Entity) {
				continue
			}
			r.renderHealthBar(b)
			visited.Set(b.Entity)
			/*text := fmt.Sprintf("%v", b.String())
			if alive := behaviors.GetAlive(b.Entity); alive != nil {
				text = fmt.Sprintf("%v", alive.Health)
			}
			r.Print(ts, int(scr[0]), int(scr[1])-16, text)*/
		}
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
		r.renderSelectedTarget()
	}

	index := 0
	for _, e := range r.Carrier.Slots {
		if e == 0 {
			continue
		}
		slot := inventory.GetSlot(e)
		if slot == nil {
			continue
		}
		r.BitBlt(slot.Image, index*40+10, r.ScreenHeight-42, 32, 32, concepts.BlendNormal)
		r.Print(ts, index*40+16+10, r.ScreenHeight-50, strconv.Itoa(slot.Count.Now))
		if e == r.Carrier.SelectedWeapon {
			r.renderWeapon(slot)
		}
		index++
	}
}

func (r *Renderer) renderHealthBar(b *core.Body) {
	// TODO: Profile memory usage here. How many of these allocations escape
	// to the heap?
	alive := behaviors.GetAlive(b.Entity)
	if alive == nil {
		return
	}
	top := &concepts.Vector3{}
	top[0] = b.Pos.Render[0]
	top[1] = b.Pos.Render[1]
	top[2] = b.Pos.Render[2] + b.Size.Render[1]*0.5
	scr := r.WorldToScreen(top)
	if scr == nil {
		return
	}

	yStart := int(scr[1]) - 7
	yEnd := int(scr[1]) - 3
	xStart := int(scr[0]) - 10
	xEnd := int(scr[0]) + 11
	// TODO: Should these be in a constants file?
	// TODO: Maybe we should load these kinds of style constants from a YAML
	// file to enable more modding.
	edge := &concepts.Vector4{0, 0, 0, 1}
	blank := &concepts.Vector4{0, 0, 0, 0.5}
	health := &concepts.Vector4{0, 1, 0, 1}

	alive.Tint(edge, &concepts.Vector4{1, 1, 1, 1})
	alive.Tint(blank, &concepts.Vector4{1, 1, 1, 1})

	switch {
	case alive.Health >= 33 && alive.Health < 66:
		health[0] = 1
		health[1] = 1
		health[2] = 0
	case alive.Health < 33:
		health[0] = 1
		health[1] = 0
		health[2] = 0
	}

	for y := yStart; y < yEnd; y++ {
		if y < 0 || y >= r.ScreenHeight {
			continue
		}
		for x := xStart; x < xEnd; x++ {
			if x < 0 || x >= r.ScreenWidth {
				continue
			}
			pixel := &r.FrameBuffer[x+y*r.ScreenWidth]
			switch {
			case x == xStart || y == yStart || x == xEnd-1 || y == yEnd-1:
				concepts.BlendColors(pixel, edge, 1.0)
			case x < xStart+int(alive.Health*float64(xEnd-xStart)/100.0):
				concepts.BlendColors(pixel, health, 1.0)
			default:
				concepts.BlendColors(pixel, blank, 1.0)
			}
		}
	}
}

func (r *Renderer) DebugInfo() {
	if r.Player == nil {
		return
	}
	//defer concepts.ExecutionDuration(concepts.ExecutionTrack("DebugInfo"))

	playerAlive := behaviors.GetAlive(r.Player.Entity)
	playerMobile := core.GetMobile(r.Player.Entity)

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

	// All visible bodies
	/*visited := ecs.EntityTable{}
	for i := range r.Blocks {
		block := &r.Blocks[i]
		for b := range block.Bodies {
			if visited.Contains(b.Entity) {
				continue
			}
			visited.Set(b.Entity)
			top := &concepts.Vector3{}
			top[0] = b.Pos.Render[0]
			top[1] = b.Pos.Render[1]
			top[2] = b.Pos.Render[2] + b.Size.Render[1]*0.5
			scr := r.WorldToScreen(top)
			if scr == nil {
				return
			}
			text := fmt.Sprintf("%v", b.String())
			if alive := behaviors.GetAlive(b.Entity); alive != nil {
				text = fmt.Sprintf("%v", alive.Health)
			}
			r.Print(ts, int(scr[0]), int(scr[1])-16, text)
		}
	}*/

	ts.HAnchor = -1
	ts.VAnchor = -1

	bodiesPerBlock := 0
	for _, block := range r.Blocks {
		bodiesPerBlock += len(block.Bodies)
	}
	r.Print(ts, 4, 4, fmt.Sprintf("FPS: %.1f, Total Entities: %v, BodiesPerBlock: %.1f", ecs.Simulation.FPS, ecs.Entities.Count(), float64(bodiesPerBlock)/float64(len(r.Blocks))))
	r.Print(ts, 4, 14, fmt.Sprintf("Health: %.1f", playerAlive.Health))
	switch 1 {
	case 0:
		hits := ecs.ComponentTableHit.Load()
		misses := ecs.ComponentTableMiss.Load()
		r.Print(ts, 4, 24, fmt.Sprintf("ComponentTable hit percentage: %.1f, %v, %v", float64(hits)*100.0/float64(hits+misses), hits, misses))
	case 1:
		tests := LightSamplerLightsTested.Load()
		total := LightSamplerCalcs.Load()
		r.Print(ts, 4, 24, fmt.Sprintf("LightSampler average lights/sample: %.1f, %v, %v", float64(tests)/float64(total), tests, total))
	}
	if r.PlayerBody.SectorEntity != 0 {
		entity := r.PlayerBody.SectorEntity
		sector := core.GetSector(entity)
		s := sector.Lightmap.Size()
		r.Print(ts, 4, 34, fmt.Sprintf("Sector: %v, LM:%v, Bodies: %v", entity.Format(), s, len(sector.Bodies)))
		r.Print(ts, 4, 44, fmt.Sprintf("f: %v, v: %v, p: %v\n", playerMobile.Force.StringHuman(2), playerMobile.Vel.Render.StringHuman(2), r.PlayerBody.Pos.Render.StringHuman(2)))
	}

	for i := range 20 {
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
