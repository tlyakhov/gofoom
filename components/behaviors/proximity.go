// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"fmt"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"

	"github.com/puzpuzpuz/xsync/v3"
	"github.com/spf13/cast"
)

//go:generate go run github.com/dmarkham/enumer -type=ProximityFlags -json
type ProximityFlags int

const (
	ProximityOnBody ProximityFlags = 1 << iota
	ProximityOnSector
	ProximityTargetsBody
	ProximityTargetsSector
)

type Proximity struct {
	ecs.Attached `editable:"^"`

	ActsOnSectors bool    `editable:"Acts on Sectors?"`
	Range         float64 `editable:"Range"`
	Hysteresis    float64 `editable:"Hysteresis (ms)"`

	ValidComponents containers.Set[ecs.ComponentID] `editable:"ValidComponents"`

	InRange core.Script `editable:"InRange"`
	Enter   core.Script `editable:"Enter"`
	Exit    core.Script `editable:"Exit"`

	State *xsync.MapOf[uint64, *ProximityState]
}

func (p *Proximity) MultiAttachable() bool { return true }

func (p *Proximity) OnAttach() {
	p.Attached.OnAttach()
	p.InRange.OnAttach()
	p.Enter.OnAttach()
	p.Exit.OnAttach()
}

func (p *Proximity) String() string {
	return fmt.Sprintf("Proximity: %.2f", p.Range)
}

func (p *Proximity) Construct(data map[string]any) {
	p.Attached.Construct(data)
	p.Range = 100
	p.Hysteresis = 200
	p.ActsOnSectors = false
	p.ValidComponents = make(containers.Set[ecs.ComponentID])
	// TODO: Serialize this
	p.State = xsync.NewMapOf[uint64, *ProximityState]()

	if data == nil {
		p.InRange.Construct(nil)
		p.Enter.Construct(nil)
		p.Exit.Construct(nil)
		return
	}

	if v, ok := data["Range"]; ok {
		p.Range = cast.ToFloat64(v)
	}

	if v, ok := data["Hysteresis"]; ok {
		p.Hysteresis = cast.ToFloat64(v)
	}

	if v, ok := data["ActsOnSectors"]; ok {
		p.ActsOnSectors = cast.ToBool(v)
	}

	if v, ok := data["InRange"]; ok {
		p.InRange.Construct(v.(map[string]any))
	}
	if v, ok := data["Enter"]; ok {
		p.Enter.Construct(v.(map[string]any))
	}
	if v, ok := data["Exit"]; ok {
		p.Exit.Construct(v.(map[string]any))
	}

	if v, ok := data["ValidComponents"]; ok {
		p.ValidComponents = ecs.ParseComponentIDs(cast.ToString(v))
	}
}

func (p *Proximity) Serialize() map[string]any {
	result := p.Attached.Serialize()

	if p.Range != 100 {
		result["Range"] = p.Range
	}
	if p.Hysteresis != 200 {
		result["Hysteresis"] = p.Hysteresis
	}

	if p.ActsOnSectors {
		result["ActsOnSectors"] = true
	}

	if !p.InRange.IsEmpty() {
		result["InRange"] = p.InRange.Serialize()
	}
	if !p.Enter.IsEmpty() {
		result["Enter"] = p.Enter.Serialize()
	}
	if !p.Exit.IsEmpty() {
		result["Exit"] = p.Exit.Serialize()
	}

	if len(p.ValidComponents) > 0 {
		result["ValidComponents"] = ecs.SerializeComponentIDs(p.ValidComponents)
	}

	return result
}
