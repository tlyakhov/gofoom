// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"strconv"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

//go:generate go run github.com/dmarkham/enumer -type=MaterialShadow -json
type MaterialShadow int

const (
	ShadowNone MaterialShadow = iota
	ShadowImage
	ShadowSphere
	ShadowAABB
)

type Visible struct {
	ecs.Attached `editable:"^"`

	Opacity   float64        `editable:"Opacity"`
	Shadow    MaterialShadow `editable:"Shadow Type"`
	PixelOnly bool           `editable:"Pixel only?"`
}

var VisibleCID ecs.ComponentID

func init() {
	VisibleCID = ecs.RegisterComponent(&ecs.Column[Visible, *Visible]{Getter: GetVisible})
}

func GetVisible(db *ecs.ECS, e ecs.Entity) *Visible {
	if asserted, ok := db.Component(e, VisibleCID).(*Visible); ok {
		return asserted
	}
	return nil
}

func (v *Visible) MultiAttachable() bool { return true }

func (v *Visible) String() string {
	return "Visible (Opacity: " + strconv.FormatFloat(v.Opacity, 'f', 2, 64) + ")"
}

func (m *Visible) Construct(data map[string]any) {
	m.Attached.Construct(data)

	m.Opacity = 1
	m.Shadow = ShadowNone
	m.PixelOnly = false

	if data == nil {
		return
	}

	if v, ok := data["Opacity"]; ok {
		m.Opacity = cast.ToFloat64(v)
	}
	if v, ok := data["PixelOnly"]; ok {
		m.PixelOnly = v.(bool)
	}
	if v, ok := data["Shadow"]; ok {
		c, err := MaterialShadowString(v.(string))
		if err == nil {
			m.Shadow = c
		} else {
			panic(err)
		}
	}
}

func (v *Visible) Serialize() map[string]any {
	result := v.Attached.Serialize()
	result["Ambient"] = v.Opacity
	if v.Shadow != ShadowNone {
		result["Shadow"] = v.Shadow.String()
	}
	if v.PixelOnly {
		result["PixelOnly"] = v.PixelOnly
	}
	return result
}
