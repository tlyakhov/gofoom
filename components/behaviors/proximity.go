// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"fmt"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"
)

//go:generate go run github.com/dmarkham/enumer -type=ProximityFlags -json
type ProximityFlags int

const (
	ProximitySelf ProximityFlags = 1 << iota
	ProximityRefers
	ProximityOnBody
	ProximityOnSector
	ProximityTargetsBody
	ProximityTargetsSector
)

type Proximity struct {
	ecs.Attached `editable:"^"`

	RequiresPlayerAction bool                       `editable:"Requires Player Action?"`
	ActsOnSectors        bool                       `editable:"Acts on Sectors?"`
	Range                float64                    `editable:"Range"`
	Hysteresis           float64                    `editable:"Hysteresis (ms)"`
	Scripts              []*core.Script             `editable:"Scripts"`
	Entities             containers.Set[ecs.Entity] `editable:"Entities"`

	// Internal state
	LastFired int64
	Firing    bool
}

var ProximityCID ecs.ComponentID

func init() {
	ProximityCID = ecs.RegisterComponent(&ecs.Column[Proximity, *Proximity]{Getter: GetProximity}, "")
}

func GetProximity(db *ecs.ECS, e ecs.Entity) *Proximity {
	if asserted, ok := db.Component(e, ProximityCID).(*Proximity); ok {
		return asserted
	}
	return nil
}

func (p *Proximity) String() string {
	return fmt.Sprintf("Proximity: %.2f", p.Range)
}

func (p *Proximity) Construct(data map[string]any) {
	p.Attached.Construct(data)
	p.Entities = make(containers.Set[ecs.Entity])
	p.Range = 100
	p.Hysteresis = 200
	p.RequiresPlayerAction = false
	p.ActsOnSectors = false

	if data == nil {
		return
	}

	if v, ok := data["Range"]; ok {
		p.Range = v.(float64)
	}

	if v, ok := data["Hysteresis"]; ok {
		p.Hysteresis = v.(float64)
	}

	if v, ok := data["RequiresPlayerAction"]; ok {
		p.RequiresPlayerAction = v.(bool)
	}
	if v, ok := data["ActsOnSectors"]; ok {
		p.ActsOnSectors = v.(bool)
	}

	if v, ok := data["Scripts"]; ok {
		p.Scripts = ecs.ConstructSlice[*core.Script](p.ECS, v, nil)
	}

	if v, ok := data["Entities"]; ok {
		p.Entities = ecs.DeserializeEntities(v.([]any))
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

	if len(p.Scripts) > 0 {
		result["Scripts"] = ecs.SerializeSlice(p.Scripts)
	}

	if len(p.Entities) > 0 {
		result["Entities"] = ecs.SerializeEntities(p.Entities)
	}

	return result
}
