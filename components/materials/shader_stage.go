// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"strconv"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

//go:generate go run github.com/dmarkham/enumer -type=ShaderFlags -json
type ShaderFlags int

const (
	ShaderTiled ShaderFlags = 1 << iota
	ShaderSky
	ShaderStaticBackground
	ShaderLiquid
	ShaderFrob
)

type ShaderStage struct {
	Universe  *ecs.Universe
	Material  ecs.Entity         `editable:"Material" edit_type:"Material"`
	Transform concepts.Matrix2   `editable:"ℝ²→ℝ²"`
	Flags     ShaderFlags        `editable:"Flags" edit_type:"Flags"`
	Frame     int                `editable:"Frame"`
	Opacity   float64            `editable:"Opacity"`
	BlendFunc concepts.BlendType `editable:"Blend"`

	IgnoreSurfaceTransform bool `editable:"Ignore Surface Transform"`
}

func (s *ShaderStage) Construct(data map[string]any) {
	s.Transform = concepts.IdentityMatrix2
	s.Flags = ShaderTiled
	s.Opacity = 1
	s.BlendFunc = concepts.BlendNormal

	if data == nil {
		return
	}

	if v, ok := data["Texture"]; ok {
		s.Material, _ = ecs.ParseEntity(v.(string))
	}

	if v, ok := data["Material"]; ok {
		s.Material, _ = ecs.ParseEntity(v.(string))
	}

	if v, ok := data["Transform"]; ok {
		s.Transform.Deserialize(v.(string))
	}

	if v, ok := data["IgnoreSurfaceTransform"]; ok {
		s.IgnoreSurfaceTransform = v.(bool)
	}

	if v, ok := data["Frame"]; ok {
		if v2, err := strconv.Atoi(v.(string)); err != nil {
			s.Frame = v2
		}
	}

	if v, ok := data["Opacity"]; ok {
		s.Opacity = cast.ToFloat64(v)
	}

	if v, ok := data["BlendingFunc"]; ok {
		s.BlendFunc, _ = concepts.BlendFuncString(cast.ToString(v))
	}

	if v, ok := data["Flags"]; ok {
		s.Flags = concepts.ParseFlags(cast.ToString(v), ShaderFlagsString)
	}
}

func (s *ShaderStage) Serialize() map[string]any {
	result := make(map[string]any)

	if s.Material != 0 {
		result["Material"] = s.Material.Serialize(s.Universe)
	}

	if s.Frame != 0 {
		result["Frame"] = strconv.Itoa(s.Frame)
	}

	if s.Opacity != 1 {
		result["Opacity"] = s.Opacity
	}

	result["BlendingFunc"] = s.BlendFunc.String()

	if s.Flags != ShaderTiled {
		result["Flags"] = concepts.SerializeFlags(s.Flags, ShaderFlagsValues())
	}
	if !s.Transform.IsIdentity() {
		result["Transform"] = s.Transform.Serialize()
	}
	if s.IgnoreSurfaceTransform {
		result["IgnoreSurfaceTransform"] = s.IgnoreSurfaceTransform
	}
	return result
}

func (s *ShaderStage) OnAttach(u *ecs.Universe) {
	s.Universe = u
}

func (s *ShaderStage) GetUniverse() *ecs.Universe {
	return s.Universe
}
