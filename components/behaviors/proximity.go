// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"fmt"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
)

type Proximity struct {
	ecs.Attached `editable:"^"`
	Range        float64        `editable:"Range"`
	Scripts      []*core.Script `editable:"Scripts"`
}

var ProximityComponentIndex int

func init() {
	ProximityComponentIndex = ecs.Types().Register(Proximity{}, ProximityFromDb)
}

func ProximityFromDb(db *ecs.ECS, e ecs.Entity) *Proximity {
	if asserted, ok := db.Component(e, ProximityComponentIndex).(*Proximity); ok {
		return asserted
	}
	return nil
}

func (p *Proximity) String() string {
	return fmt.Sprintf("Proximity: %.2f", p.Range)
}

func (p *Proximity) Construct(data map[string]any) {
	p.Attached.Construct(data)

	p.Range = 100

	if data == nil {
		return
	}

	if v, ok := data["Range"]; ok {
		p.Range = v.(float64)
	}

	if v, ok := data["Scripts"]; ok {
		p.Scripts = ecs.ConstructSlice[*core.Script](p.ECS, v)
	}

}

func (p *Proximity) Serialize() map[string]any {
	result := p.Attached.Serialize()
	result["Range"] = p.Range

	if len(p.Scripts) > 0 {
		result["Scripts"] = ecs.SerializeSlice(p.Scripts)
	}

	return result
}
