// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"tlyakhov/gofoom/ecs"

	"github.com/gammazero/deque"
	"github.com/spf13/cast"
)

type Mark struct {
	*ShaderStage
	*Surface
}

type MarkMaker struct {
	ecs.Attached `editable:"^"`
	// Projectiles make marks on walls/internal segments
	Material ecs.Entity `editable:"Material" edit_type:"Material"`
	Size     float64    `editable:"Size"`

	// TODO: We should serialize these
	Marks deque.Deque[Mark]
}

func (m *MarkMaker) Construct(data map[string]any) {
	m.Attached.Construct(data)
	m.Marks = deque.Deque[Mark]{}
	m.Size = 5

	if data == nil {
		return
	}

	if v, ok := data["Material"]; ok {
		m.Material, _ = ecs.ParseEntity(v.(string))
	}

	if v, ok := data["Size"]; ok {
		m.Size = cast.ToFloat64(v)
	}

}

func (m *MarkMaker) Serialize() map[string]any {
	result := m.Attached.Serialize()
	if m.Material != 0 {
		result["Material"] = m.Material.Serialize()
	}

	if m.Size != 5 {
		result["Size"] = m.Size
	}

	return result
}
