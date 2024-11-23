// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"fmt"
	"strings"
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

	RequiresPlayerAction bool    `editable:"Requires Player Action?"`
	ActsOnSectors        bool    `editable:"Acts on Sectors?"`
	Range                float64 `editable:"Range"`
	Hysteresis           float64 `editable:"Hysteresis (ms)"`

	ValidComponents containers.Set[ecs.ComponentID] `editable:"ValidComponents"`

	InRange core.Script `editable:"InRange"`
	Enter   core.Script `editable:"Enter"`
	Exit    core.Script `editable:"Exit"`

	State *xsync.MapOf[uint64, *ProximityState]
}

var ProximityCID ecs.ComponentID

func init() {
	ProximityCID = ecs.RegisterComponent(&ecs.Column[Proximity, *Proximity]{Getter: GetProximity})
}

func GetProximity(db *ecs.ECS, e ecs.Entity) *Proximity {
	if asserted, ok := db.Component(e, ProximityCID).(*Proximity); ok {
		return asserted
	}
	return nil
}

func (p *Proximity) MultiAttachable() bool { return true }

func (p *Proximity) AttachECS(db *ecs.ECS) {
	p.Attached.AttachECS(db)
	p.InRange.AttachECS(db)
	p.Enter.AttachECS(db)
	p.Exit.AttachECS(db)
}

func (p *Proximity) String() string {
	return fmt.Sprintf("Proximity: %.2f", p.Range)
}

func (p *Proximity) Construct(data map[string]any) {
	p.Attached.Construct(data)
	p.Range = 100
	p.Hysteresis = 200
	p.RequiresPlayerAction = false
	p.ActsOnSectors = false
	p.ValidComponents = make(containers.Set[ecs.ComponentID])
	// TODO: Serialize this
	p.State = xsync.NewMapOf[uint64, *ProximityState]()

	if data == nil {
		return
	}

	if v, ok := data["Range"]; ok {
		p.Range = cast.ToFloat64(v)
	}

	if v, ok := data["Hysteresis"]; ok {
		p.Hysteresis = cast.ToFloat64(v)
	}

	if v, ok := data["RequiresPlayerAction"]; ok {
		p.RequiresPlayerAction = v.(bool)
	}
	if v, ok := data["ActsOnSectors"]; ok {
		p.ActsOnSectors = v.(bool)
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
		split := strings.Split(v.(string), ",")
		for _, s := range split {
			id := ecs.Types().IDs[s]
			if id != 0 {
				p.ValidComponents.Add(id)
			}
		}
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

	if p.RequiresPlayerAction {
		result["RequiresPlayerAction"] = true
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
		s := ""
		for id := range p.ValidComponents {
			if len(s) > 0 {
				s += ","
			}
			s += ecs.Types().ColumnPlaceholders[id].String()
		}
		result["ValidComponents"] = s
	}

	return result
}
