// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"strconv"
	"tlyakhov/gofoom/concepts"
)

type Sprite struct {
	concepts.Attached `editable:"^"`

	Image  concepts.EntityRef[*Image] `editable:"Image" edit_type:"Material"`
	Cols   uint32                     `editable:"Columns"`
	Rows   uint32                     `editable:"Rows"`
	Angles uint32                     `editable:"# of Angles"`
}

var SpriteComponentIndex int

func init() {
	SpriteComponentIndex = concepts.DbTypes().Register(Sprite{}, SpriteFromDb)
}

func SpriteFromDb(db *concepts.EntityComponentDB, e concepts.Entity) *Sprite {
	if asserted, ok := db.Component(e, SpriteComponentIndex).(*Sprite); ok {
		return asserted
	}
	return nil
}

func (s *Sprite) Sample(x, y float64, sw, sh uint32, c, r uint32) concepts.Vector4 {
	if s.Image.Value == nil || s.Rows == 0 || s.Cols == 0 {
		// Debug values
		return concepts.Vector4{x, 0, y, 1} // full alpha
	}

	if x < 0 || y < 0 || x >= 1 || y >= 1 {
		return concepts.Vector4{0, 0, 0, 0}
	}

	x = (float64(c) + x) / float64(s.Cols)
	y = (float64(r) + y) / float64(s.Rows)

	return s.Image.Value.Sample(x, y, sw*uint32(s.Cols), sh*uint32(s.Rows))
}

func (s *Sprite) SampleAlpha(x, y float64, sw, sh uint32, c, r uint32) float64 {
	if s.Image.Value == nil || s.Rows == 0 || s.Cols == 0 {
		return 1 // full alpha
	}

	if x < 0 || y < 0 || x >= 1 || y >= 1 {
		return 0
	}

	x = (float64(c) + x) / float64(s.Cols)
	y = (float64(r) + y) / float64(s.Rows)

	return s.Image.Value.SampleAlpha(x, y, sw*uint32(s.Cols), sh*uint32(s.Rows))
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
		s.Image.Entity, _ = concepts.ParseEntity(v.(string))
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
	if s.Image.Entity != 0 {
		result["Image"] = s.Image.Entity.Format()
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
