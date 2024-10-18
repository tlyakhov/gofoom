// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"strconv"
	"tlyakhov/gofoom/containers"
)

// Adding an Instanced component to an entity makes it possible for that entity
// to mix in components from other entities, reusing common data. This does
// incur a cost when accessing, since the component table has to follow the
// chain of references, and controllers have to substitute entity IDs for the
// instanced components. It also adds complexity for any code that relies on
// <Component>.Entity to be consistent.
type Instanced struct {
	Attached `editable:"^"`
	// TODO: Should these be an ordered list for layering?
	Sources containers.Set[Entity] `editable:"Source Entities"`

	// Internal state
	SourceComponents ComponentTable
}

var InstancedCID ComponentID

func init() {
	InstancedCID = RegisterComponent(&Column[Instanced, *Instanced]{Getter: GetInstanced}, "")
}

func GetInstanced(db *ECS, e Entity) *Instanced {
	if asserted, ok := db.Component(e, InstancedCID).(*Instanced); ok {
		return asserted
	}
	return nil
}

func (n *Instanced) String() string {
	return strconv.FormatInt(int64(n.Entity), 10)
}

func (n *Instanced) Recalculate() {
	for _, sourceComponent := range n.SourceComponents {
		if sourceComponent != nil {
			sourceComponent.Base().instancedCopies.Delete(n.Entity)
		}
	}
	n.SourceComponents = make(ComponentTable, 0)
	for sourceEntity := range n.Sources {
		for _, sourceComponent := range n.ECS.AllComponents(sourceEntity) {
			if sourceComponent == nil {
				continue
			}
			n.SourceComponents.Set(sourceComponent)
			sourceComponent.Base().instancedCopies.Add(n.Entity)
		}
	}
}

func (n *Instanced) PushEntityFields() {
	for sourceEntity := range n.Sources {
		for _, sourceComponent := range n.ECS.AllComponents(sourceEntity) {
			if sourceComponent == nil {
				continue
			}
			a := sourceComponent.Base()
			a.entityStack = append(a.entityStack, a.Entity)
			a.Entity = n.Entity
		}
	}
}

func (n *Instanced) PopEntityFields() {
	for sourceEntity := range n.Sources {
		for _, sourceComponent := range n.ECS.AllComponents(sourceEntity) {
			if sourceComponent == nil {
				continue
			}
			a := sourceComponent.Base()
			last := len(a.entityStack) - 1
			a.Entity = a.entityStack[last]
			a.entityStack = a.entityStack[:last]
		}
	}
}

func (n *Instanced) Construct(data map[string]any) {
	n.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Entities"]; ok {
		n.Sources = DeserializeEntities(v.([]any))
	}
}

func (n *Instanced) Serialize() map[string]any {
	result := n.Attached.Serialize()
	result["Entities"] = SerializeEntities(n.Sources)
	return result
}
