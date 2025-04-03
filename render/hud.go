// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

type playerMessageParams struct {
	TargetableEntity ecs.Entity
	PlayerTargetable *behaviors.PlayerTargetable
	Player           *character.Player
	Carrier          *inventory.Carrier
}

func (r *Renderer) RenderWeapon(slot *inventory.Slot) {
	wc := inventory.GetWeaponClass(slot.Universe, slot.Entity)
	wi := inventory.GetWeapon(slot.Universe, slot.Entity)
	if wc != nil && wi != nil {
		if wi.Flashing() {
			r.BitBlt(wc.FlashMaterial, r.ScreenWidth/2-64, r.ScreenHeight-160, 128, 128, concepts.BlendScreen)
		}
	}
	// TODO: This should be a separate image from the inventory item image
	r.BitBlt(slot.Image, r.ScreenWidth/2-64, r.ScreenHeight-128, 128, 128, concepts.BlendNormal)
}

func applyPlayerMessage(pt *behaviors.PlayerTargetable, params *playerMessageParams) string {
	if pt.MessageTemplate == nil {
		return pt.Message
	}

	var buf bytes.Buffer
	err := pt.MessageTemplate.Execute(&buf, params)
	if err != nil {
		return fmt.Sprintf("Error in message template %v: %v", pt.Message, err)
	}
	return buf.String()
}

func (r *Renderer) renderSelectedTarget() {
	ts := r.textStyle

	pt := behaviors.GetPlayerTargetable(r.Universe, r.Player.SelectedTarget)
	if pt == nil || len(pt.Message) == 0 {
		return
	}

	scr := r.WorldToScreen(pt.Pos(r.Player.SelectedTarget))
	if scr == nil {
		return
	}

	params := &playerMessageParams{
		TargetableEntity: r.Player.SelectedTarget,
		PlayerTargetable: pt,
		Player:           r.Player,
		Carrier:          r.Carrier,
	}
	msg := strings.TrimSpace(applyPlayerMessage(pt, params))
	r.Print(ts, int(scr[0]), int(scr[1])-16, msg)
}

func (r *Renderer) RenderHud() {
	if r.Player == nil || r.Carrier == nil {
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
		r.renderSelectedTarget()
	}

	index := 0
	for _, e := range r.Carrier.Slots {
		if e == 0 {
			continue
		}
		slot := inventory.GetSlot(r.Universe, e)
		if slot == nil {
			continue
		}
		r.BitBlt(slot.Image, index*40+10, r.ScreenHeight-42, 32, 32, concepts.BlendNormal)
		r.Print(ts, index*40+16+10, r.ScreenHeight-50, strconv.Itoa(slot.Count.Now))
		if e == r.Carrier.SelectedWeapon {
			r.RenderWeapon(slot)
		}
		index++
	}
}

func (r *Renderer) DebugInfo() {
	//defer concepts.ExecutionDuration(concepts.ExecutionTrack("DebugInfo"))

	playerAlive := behaviors.GetAlive(r.Universe, r.Player.Entity)
	playerMobile := core.GetMobile(r.Universe, r.Player.Entity)

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
			if alive := behaviors.GetAlive(r.Universe, b.Entity); alive != nil {
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
	r.Print(ts, 4, 4, fmt.Sprintf("FPS: %.1f, Total Entities: %v, BodiesPerBlock: %.1f", r.Universe.Simulation.FPS, r.Universe.Entities.Count(), float64(bodiesPerBlock)/float64(len(r.Blocks))))
	r.Print(ts, 4, 14, fmt.Sprintf("Health: %.1f", playerAlive.Health))
	switch 2 {
	case 0:
		hits := r.ICacheHits.Load()
		misses := r.ICacheMisses.Load()
		r.Print(ts, 4, 24, fmt.Sprintf("ICache hit percentage: %.1f, %v, %v", float64(hits)*100.0/float64(hits+misses), hits, misses))
	case 1:
		hits := ecs.ComponentTableHit.Load()
		misses := ecs.ComponentTableMiss.Load()
		r.Print(ts, 4, 24, fmt.Sprintf("ComponentTable hit percentage: %.1f, %v, %v", float64(hits)*100.0/float64(hits+misses), hits, misses))
	case 2:
		tests := LightSamplerLightsTested.Load()
		total := LightSamplerCalcs.Load()
		r.Print(ts, 4, 24, fmt.Sprintf("LightSampler average lights/sample: %.1f, %v, %v", float64(tests)/float64(total), tests, total))
	}
	if r.PlayerBody.SectorEntity != 0 {
		entity := r.PlayerBody.SectorEntity
		sector := core.GetSector(r.Universe, entity)
		s := sector.Lightmap.Size()
		r.Print(ts, 4, 34, fmt.Sprintf("Sector: %v, LM:%v, Bodies: %v", entity.Format(r.Universe), s, len(sector.Bodies)))
		r.Print(ts, 4, 44, fmt.Sprintf("f: %v, v: %v, p: %v\n", playerMobile.Force.StringHuman(2), playerMobile.Vel.Render.StringHuman(2), r.PlayerBody.Pos.Render.StringHuman(2)))
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
