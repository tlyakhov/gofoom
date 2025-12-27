// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"strconv"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type SpriteSheet struct {
	ecs.Attached `editable:"^"`

	Material ecs.Entity `editable:"Material" edit_type:"Material"`
	Cols     uint32     `editable:"Columns"`
	Rows     uint32     `editable:"Rows"`
	Angles   uint32     `editable:"# of Angles"`

	Frames uint32                    `editable:"# of Frames"`
	Frame  dynamic.DynamicValue[int] `editable:"Frame"`
}

func (s *SpriteSheet) Shareable() bool { return true }

func (s *SpriteSheet) OnDelete() {
	defer s.Attached.OnDelete()
	if s.IsAttached() {
		s.Frame.Detach(ecs.Simulation)
	}
}

func (s *SpriteSheet) OnAttach() {
	s.Attached.OnAttach()
	s.Frame.Attach(ecs.Simulation)

}

func (s *SpriteSheet) TransformUV(u, v float64, c, r uint32) (ur, vr float64) {
	if s.Material == 0 || s.Rows == 0 || s.Cols == 0 {
		return u, v
	}

	if u < 0 || v < 0 || u >= 1 || v >= 1 {
		return -1, -1
	}

	return (float64(c) + u) / float64(s.Cols), (float64(r) + v) / float64(s.Rows)
}

func (s *SpriteSheet) Construct(data map[string]any) {
	s.Attached.Construct(data)

	s.Rows = 1
	s.Cols = 1
	s.Angles = 1
	s.Frames = 1
	s.Frame.Construct(nil)

	if data == nil {
		return
	}

	if v, ok := data["Material"]; ok {
		s.Material, _ = ecs.ParseEntity(v.(string))
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
		s.Frame.Construct(v)
	}
}

func (s *SpriteSheet) Serialize() map[string]any {
	result := s.Attached.Serialize()
	if s.Material != 0 {
		result["Material"] = s.Material.Serialize()
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
