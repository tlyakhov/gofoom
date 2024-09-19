// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"strconv"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type Sprite struct {
	ecs.Attached `editable:"^"`

	Image  ecs.Entity `editable:"Image" edit_type:"Material"`
	Cols   uint32     `editable:"Columns"`
	Rows   uint32     `editable:"Rows"`
	Angles uint32     `editable:"# of Angles"`

	Frames uint32                    `editable:"# of Frames"`
	Frame  dynamic.DynamicValue[int] `editable:"Frame"`
}

var SpriteCID ecs.ComponentID

func init() {
	SpriteCID = ecs.RegisterComponent(&ecs.Column[Sprite, *Sprite]{Getter: GetSprite}, "")
}

func GetSprite(db *ecs.ECS, e ecs.Entity) *Sprite {
	if asserted, ok := db.Component(e, SpriteCID).(*Sprite); ok {
		return asserted
	}
	return nil
}

func (s *Sprite) OnDetach() {
	if s.ECS != nil {
		s.Frame.Detach(s.ECS.Simulation)
	}
	s.Attached.OnDetach()
}

func (s *Sprite) AttachECS(db *ecs.ECS) {
	if s.ECS != db {
		s.OnDetach()
	}
	s.Attached.AttachECS(db)
	s.Frame.Attach(db.Simulation)

}

func (s *Sprite) TransformUV(u, v float64, c, r uint32) (ur, vr float64) {
	if s.Image == 0 || s.Rows == 0 || s.Cols == 0 {
		return u, v
	}

	if u < 0 || v < 0 || u >= 1 || v >= 1 {
		return -1, -1
	}

	return (float64(c) + u) / float64(s.Cols), (float64(r) + v) / float64(s.Rows)
}

func (s *Sprite) Construct(data map[string]any) {
	s.Attached.Construct(data)

	s.Rows = 1
	s.Cols = 1
	s.Angles = 1
	s.Frames = 1
	s.Frame.Construct(nil)

	if data == nil {
		return
	}

	if v, ok := data["Image"]; ok {
		s.Image, _ = ecs.ParseEntity(v.(string))
	}

	if v, ok := data["Rows"]; ok {
		if v2, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
			s.Rows = uint32(v2)
		}
	}
	if v, ok := data["Cols"]; ok {
		if v2, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
			s.Cols = uint32(v2)
		}
	}
	if v, ok := data["Angles"]; ok {
		if v2, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
			s.Angles = uint32(v2)
		}
	}
	if v, ok := data["Frames"]; ok {
		if v2, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
			s.Frames = uint32(v2)
		}
	}
	if v, ok := data["Frame"]; ok {
		s.Frame.Construct(v.(map[string]any))
	}
}

func (s *Sprite) Serialize() map[string]any {
	result := s.Attached.Serialize()
	if s.Image != 0 {
		result["Image"] = s.Image.String()
	}
	if s.Rows != 1 {
		result["Rows"] = strconv.FormatUint(uint64(s.Rows), 10)
	}
	if s.Cols != 1 {
		result["Cols"] = strconv.FormatUint(uint64(s.Cols), 10)
	}
	if s.Angles != 1 {
		result["Angles"] = strconv.FormatUint(uint64(s.Angles), 10)
	}
	if s.Frames != 1 {
		result["Frames"] = strconv.FormatUint(uint64(s.Frames), 10)
	}
	result["Frame"] = s.Frame.Serialize()

	return result
}
