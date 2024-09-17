// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"strconv"
	"strings"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

//go:generate go run github.com/dmarkham/enumer -type=ShaderFlags -json
type ShaderFlags int

const (
	ShaderTiled ShaderFlags = 1 << iota
	ShaderSky
	ShaderStaticBackground
	ShaderLiquid
)

type ShaderStage struct {
	ECS       *ecs.ECS
	System    bool
	Texture   ecs.Entity       `editable:"Texture" edit_type:"Material"`
	Transform concepts.Matrix2 `editable:"Transform"`
	Flags     ShaderFlags      `editable:"Flags" edit_type:"Flags"`
	Frame     int              `editable:"Frame"`

	IgnoreSurfaceTransform bool `editable:"Ignore Surface Transform"`

	// TODO: implement
	Blend any
}

func (s *ShaderStage) Construct(data map[string]any) {
	s.Transform = concepts.IdentityMatrix2
	s.Flags = ShaderTiled

	if data == nil {
		return
	}

	if v, ok := data["Texture"]; ok {
		s.Texture, _ = ecs.ParseEntity(v.(string))
	}

	if v, ok := data["Transform"]; ok {
		s.Transform.Deserialize(v.([]any))
	}

	if v, ok := data["IgnoreSurfaceTransform"]; ok {
		s.IgnoreSurfaceTransform = v.(bool)
	}

	if v, ok := data["Frame"]; ok {
		if v2, err := strconv.Atoi(v.(string)); err != nil {
			s.Frame = v2
		}
	}

	if v, ok := data["Flags"]; ok {
		parsedFlags := strings.Split(v.(string), "|")
		s.Flags = 0
		for _, parsedFlag := range parsedFlags {
			if f, err := ShaderFlagsString(parsedFlag); err == nil {
				s.Flags |= f
			}
		}
	}
}

func (s *ShaderStage) Serialize() map[string]any {
	result := make(map[string]any)

	if s.Texture != 0 {
		result["Texture"] = s.Texture.String()
	}

	if s.Frame != 0 {
		result["Frame"] = strconv.Itoa(s.Frame)
	}

	if s.Flags != ShaderTiled {
		flags := ""
		for _, f := range ShaderFlagsValues() {
			if s.Flags&f == 0 {
				continue
			}
			if len(flags) > 0 {
				flags += "|"
			}
			flags += f.String()
		}
		result["Flags"] = flags
	}
	if !s.Transform.IsIdentity() {
		result["Transform"] = s.Transform.Serialize()
	}
	if s.IgnoreSurfaceTransform {
		result["IgnoreSurfaceTransform"] = s.IgnoreSurfaceTransform
	}
	return result
}

func (s *ShaderStage) AttachECS(db *ecs.ECS) {
	s.ECS = db
}

func (s *ShaderStage) GetECS() *ecs.ECS {
	return s.ECS
}

func (s *ShaderStage) IsSystem() bool {
	return s.System
}
