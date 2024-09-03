// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"fmt"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"
)

type Proximity struct {
	ecs.Attached `editable:"^"`
	Range        float64                    `editable:"Range"`
	Scripts      []*core.Script             `editable:"Scripts"`
	Entities     containers.Set[ecs.Entity] `editable:"Entities"`
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

	if data == nil {
		return
	}

	if v, ok := data["Range"]; ok {
		p.Range = v.(float64)
	}

	if v, ok := data["Scripts"]; ok {
		p.Scripts = ecs.ConstructSlice[*core.Script](p.ECS, v)
	}

	if v, ok := data["Entities"]; ok {
		p.Entities = ecs.DeserializeEntities(v.([]any))
	}
}

func (p *Proximity) Serialize() map[string]any {
	result := p.Attached.Serialize()
	result["Range"] = p.Range

	if len(p.Scripts) > 0 {
		result["Scripts"] = ecs.SerializeSlice(p.Scripts)
	}

	if len(p.Entities) > 0 {
		result["Entities"] = ecs.SerializeEntities(p.Entities)
	}

	return result
}
