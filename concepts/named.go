// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"github.com/rs/xid"
)

type Named struct {
	Attached `editable:"^"`
	Name     string `editable:"Name"`
}

var NamedComponentIndex int

func init() {
	NamedComponentIndex = DbTypes().Register(Named{}, NamedFromDb)
}

func NamedFromDb(db *EntityComponentDB, e Entity) *Named {
	if asserted, ok := db.Component(e, NamedComponentIndex).(*Named); ok {
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
		n.Name = v.(string)
	}
}

func (n *Named) Serialize() map[string]any {
	result := n.Attached.Serialize()
	result["Name"] = n.Name
	return result
}
