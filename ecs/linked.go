// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"strconv"
)

// Adding an Linked component to an entity makes it possible for that entity
// to mix in components from other entities, reusing common data.
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

func (n *Linked) OnDetach(e Entity) {
	defer n.Attached.OnDetach(e)
	if n.ECS == nil {
		return
	}
	// Remove this entity from any linked copies
	for _, c := range n.SourceComponents {
		if c != nil {
			n.ECS.detach(c.Base().ComponentID, n.Entity, false)
		}
	}
	n.SourceComponents = make(ComponentTable, 0)
}

func (n *Linked) Recalculate() {
	// Remove this entity from any linked copies
	for _, c := range n.SourceComponents {
		if c != nil {
			n.ECS.detach(c.Base().ComponentID, n.Entity, false)
		}
	}
	n.SourceComponents = make(ComponentTable, 0)
	for _, sourceEntity := range n.Sources {
		for _, c := range n.ECS.AllComponents(sourceEntity) {
			if c == nil || !c.MultiAttachable() {
				continue
			}
			n.SourceComponents.Set(c)
			n.ECS.attach(n.Entity, &c, c.Base().ComponentID)
		}
	}
}

func (n *Linked) Construct(data map[string]any) {
	n.Attached.Construct(data)

	n.Sources = make([]Entity, 0)

	if data == nil {
		return
	}

	if v, ok := data["Sources"]; ok {
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
	result["Sources"] = arr
	return result
}
