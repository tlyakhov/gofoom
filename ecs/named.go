// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"github.com/rs/xid"
	"github.com/spf13/cast"
)

type Named struct {
	Attached `editable:"^"`
	Name     string `editable:"Name"`
}

var NamedCID ComponentID

func init() {
	NamedCID = RegisterComponent(&Column[Named, *Named]{Getter: GetNamed})
}

func (*Named) ComponentID() ComponentID {
	return NamedCID
}

func GetNamed(u *Universe, e Entity) *Named {
	if asserted, ok := u.Component(e, NamedCID).(*Named); ok {
		return asserted
	}
	return nil
}

func (n *Named) String() string {
	return n.Name
}

func (n *Named) Construct(data map[string]any) {
	n.Attached.Construct(data)

	n.Name = xid.New().String()

	if data == nil {
		return
	}
	if v, ok := data["Name"]; ok {
		n.Name = cast.ToString(v)
	}
}

func (n *Named) Serialize() map[string]any {
	result := n.Attached.Serialize()
	result["Name"] = n.Name
	return result
}
