// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"tlyakhov/gofoom/concepts"
)

type Solid struct {
	concepts.Attached `editable:"^"`
	Diffuse           concepts.DynamicValue[concepts.Vector4] `editable:"Color"`
}

var SolidComponentIndex int

func init() {
	SolidComponentIndex = concepts.DbTypes().Register(Solid{}, SolidFromDb)
}

func SolidFromDb(db *concepts.EntityComponentDB, e concepts.Entity) *Solid {
	if asserted, ok := db.Component(e, SolidComponentIndex).(*Solid); ok {
		return asserted
	}
	return nil
}

func (s *Solid) OnDetach() {
	if s.DB != nil {
		s.Diffuse.Detach(s.DB.Simulation)
	}
	s.Attached.OnDetach()
}
func (s *Solid) SetDB(db *concepts.EntityComponentDB) {
	if s.DB != db {
		s.OnDetach()
	}
	s.Attached.SetDB(db)
	s.Diffuse.Attach(s.DB.Simulation)
}

func (s *Solid) Construct(data map[string]any) {
	s.Attached.Construct(data)
	s.Diffuse.Construct(nil)

	if data == nil {
		return
	}

	if v, ok := data["Diffuse"]; ok {
		s.Diffuse.Construct(v.(map[string]any))
	}
}

func (s *Solid) Serialize() map[string]any {
	result := s.Attached.Serialize()
	result["Diffuse"] = s.Diffuse.Serialize()
	return result
}
