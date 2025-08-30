// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"github.com/spf13/cast"
)

// Adding an Linked component to an entity makes it possible for that entity
// to mix in components from other entities, reusing common data.
type Linked struct {
	Attached `editable:"^"`
	// Ordered list, can be layered
	Sources []Entity `editable:"Source Entities"`

	AlwaysReplace bool `editable:"Always Replace?"`

	// Internal state
	SourceComponents ComponentTable
}

var LinkedCID ComponentID

func init() {
	LinkedCID = RegisterComponent(&Arena[Linked, *Linked]{})
}

func GetLinked(e Entity) *Linked {
	if asserted, ok := GetComponent(e, LinkedCID).(*Linked); ok {
		return asserted
	}
	return nil
}

func (*Linked) ComponentID() ComponentID {
	return LinkedCID
}

func (n *Linked) String() string {
	result := ""
	for _, e := range n.Sources {
		if len(result) > 0 {
			result += ", "
		}
		result += e.String()
	}
	return "Linked: " + result
}

func (n *Linked) OnDetach(e Entity) {
	defer n.Attached.OnDetach(e)
	if !n.IsAttached() {
		return
	}
	// Remove this entity from any linked copies
	for _, c := range n.SourceComponents {
		if c != nil {
			detach(c.ComponentID(), n.Entity, false)
		}
	}
	n.SourceComponents = make(ComponentTable, 0)
}

func (n *Linked) Construct(data map[string]any) {
	n.Attached.Construct(data)

	n.Sources = make([]Entity, 0)
	n.AlwaysReplace = false

	if data == nil {
		return
	}

	if v, ok := data["Sources"]; ok {
		n.Sources = ParseEntitySlice(v, true)
	}

	if v, ok := data["AlwaysReplace"]; ok {
		n.AlwaysReplace = cast.ToBool(v)
	}
}

func (n *Linked) Serialize() map[string]any {
	result := n.Attached.Serialize()
	arr := make([]string, len(n.Sources))
	for i, e := range n.Sources {
		arr[i] = e.Serialize()
	}
	result["Sources"] = arr
	if n.AlwaysReplace {
		result["AlwaysReplace"] = true
	}
	return result
}
