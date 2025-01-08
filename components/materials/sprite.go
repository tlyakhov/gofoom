// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type Sprite struct {
	ecs.Attached `editable:"^"`

	Material ecs.Entity                `editable:"Material" edit_type:"Material"`
	Frame    dynamic.DynamicValue[int] `editable:"Frame"`
}

var SpriteCID ecs.ComponentID

func init() {
	SpriteCID = ecs.RegisterComponent(&ecs.Column[Sprite, *Sprite]{Getter: GetSprite})
}

func GetSprite(db *ecs.ECS, e ecs.Entity) *Sprite {
	if asserted, ok := db.Component(e, SpriteCID).(*Sprite); ok {
		return asserted
	}
	return nil
}

func (s *Sprite) MultiAttachable() bool { return true }

func (s *Sprite) OnDelete() {
	defer s.Attached.OnDelete()
	if s.ECS != nil {
		s.Frame.Detach(s.ECS.Simulation)
	}
}

func (s *Sprite) OnAttach(db *ecs.ECS) {
	s.Attached.OnAttach(db)
	s.Frame.Attach(db.Simulation)

}

func (s *Sprite) Construct(data map[string]any) {
	s.Attached.Construct(data)

	s.Frame.Construct(nil)

	if data == nil {
		return
	}

	if v, ok := data["Material"]; ok {
		s.Material, _ = ecs.ParseEntity(v.(string))
	}

	if v, ok := data["Frame"]; ok {
		s.Frame.Construct(v.(map[string]any))
	}
}

func (s *Sprite) Serialize() map[string]any {
	result := s.Attached.Serialize()
	if s.Material != 0 {
		result["Material"] = s.Material.String()
	}
	result["Frame"] = s.Frame.Serialize()

	return result
}
