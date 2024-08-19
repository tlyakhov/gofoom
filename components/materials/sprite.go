// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"strconv"
	"tlyakhov/gofoom/ecs"
)

type Sprite struct {
	ecs.Attached `editable:"^"`

	Image  ecs.Entity `editable:"Image" edit_type:"Material"`
	Cols   uint32     `editable:"Columns"`
	Rows   uint32     `editable:"Rows"`
	Angles uint32     `editable:"# of Angles"`
}

var SpriteComponentIndex int

func init() {
	SpriteComponentIndex = ecs.Types().Register(Sprite{}, GetSprite)
}

func GetSprite(db *ecs.ECS, e ecs.Entity) *Sprite {
	if asserted, ok := db.Component(e, SpriteComponentIndex).(*Sprite); ok {
		return asserted
	}
	return nil
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
	s.Angles = 0

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
}

func (s *Sprite) Serialize() map[string]any {
	result := s.Attached.Serialize()
	if s.Image != 0 {
		result["Image"] = s.Image.Format()
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

	return result
}
