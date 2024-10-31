// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"strconv"
)

// Adding an Linked component to an entity makes it possible for that entity
// to mix in components from other entities, reusing common data. This does
// incur a cost when accessing, since the component table has to follow the
// chain of references, and controllers have to substitute entity IDs for the
// linked components. It also adds complexity for any code that relies on
// <Component>.Entity to be consistent.
type Linked struct {
	Attached `editable:"^"`
	// Ordered list, can be layered
	Sources []Entity `editable:"Source Entities"`

	// Internal state
	SourceComponents ComponentTable
}

var LinkedCID ComponentID

func init() {
	LinkedCID = RegisterComponent(&Column[Linked, *Linked]{Getter: GetLinked})
}

func GetLinked(db *ECS, e Entity) *Linked {
	if asserted, ok := db.Component(e, LinkedCID).(*Linked); ok {
		return asserted
	}
	return nil
}

func (n *Linked) String() string {
	return strconv.FormatInt(int64(n.Entity), 10)
}

func (n *Linked) OnDetach() {
	defer n.Attached.OnDetach()
	if n.ECS == nil {
		return
	}
	// Remove this entity from any linked copies
	for _, sourceComponent := range n.SourceComponents {
		if sourceComponent != nil {
			sourceComponent.Base().linkedCopies.Delete(n.UnlinkedEntity())
		}
	}
	n.SourceComponents = make(ComponentTable, 0)
}

func (n *Linked) Recalculate() {
	for _, sourceComponent := range n.SourceComponents {
		if sourceComponent != nil {
			sourceComponent.Base().linkedCopies.Delete(n.UnlinkedEntity())
		}
	}
	n.SourceComponents = make(ComponentTable, 0)
	for _, sourceEntity := range n.Sources {
		for _, sourceComponent := range n.ECS.AllComponents(sourceEntity) {
			if sourceComponent == nil {
				continue
			}
			n.SourceComponents.Set(sourceComponent)
			sourceComponent.Base().linkedCopies.Add(n.UnlinkedEntity())
		}
	}
}

// These methods recurse towards roots. In other words, if you have:
// A < B < C (where B & C are ecs.Linked components) and you do
// C.pushEntityFields, it will recurse into A.
func (n *Linked) PushEntityFields() {
	for _, sourceEntity := range n.Sources {
		for i, sourceComponent := range n.ECS.AllComponents(sourceEntity) {
			if sourceComponent == nil {
				continue
			}
			parent := sourceComponent.Base()
			parent.entityStack = append(parent.entityStack, parent.Entity)
			parent.Entity = n.Entity
			if i == 0 { // Recurse into linked component
				sourceComponent.(*Linked).PushEntityFields()
			}
		}
	}
}

func (n *Linked) PopEntityFields() {
	for _, sourceEntity := range n.Sources {
		for i, sourceComponent := range n.ECS.AllComponents(sourceEntity) {
			if sourceComponent == nil {
				continue
			}
			if i == 0 { // Recurse into linked component
				sourceComponent.(*Linked).PopEntityFields()
			}
			parent := sourceComponent.Base()
			last := len(parent.entityStack) - 1
			parent.Entity = parent.entityStack[last]
			parent.entityStack = parent.entityStack[:last]
		}
	}
}

func (n *Linked) Construct(data map[string]any) {
	n.Attached.Construct(data)

	n.Sources = make([]Entity, 0)

	if data == nil {
		return
	}

	if v, ok := data["Entities"]; ok {
		arr := v.([]any)
		n.Sources = make([]Entity, len(arr))
		for i, e := range arr {
			n.Sources[i], _ = ParseEntity(e.(string))
		}
	}
}

func (n *Linked) Serialize() map[string]any {
	result := n.Attached.Serialize()
	arr := make([]string, len(n.Sources))
	for i, e := range n.Sources {
		arr[i] = e.String()
	}
	result["Entities"] = arr
	return result
}
