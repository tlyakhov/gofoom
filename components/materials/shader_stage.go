// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"math"
	"sync"
	"sync/atomic"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

//go:generate go run github.com/dmarkham/enumer -type=ShaderFlags -json
type ShaderFlags uint32

const (
	ShaderTiled ShaderFlags = 1 << iota
	ShaderSky
	ShaderStaticBackground
	ShaderLiquid
	ShaderFrob
)

// TODO: this is a bit ugly - the stages conflate several types of operations:
//  1. Sampling - e.g. "give me a color based on this UV"
//  2. Pick source - e.g. "give me something that generates pixels"
//  3. UV transformation or replacement
//  4. Blending
//
// We should separate these ideas, there's a bunch of prior art here.
type ShaderStage struct {
	Material  ecs.Entity         `editable:"Material" edit_type:"Material"`
	Transform concepts.Matrix2   `editable:"ℝ²→ℝ²"`
	Flags     ShaderFlags        `editable:"Flags" edit_type:"Flags"`
	Frame     int                `editable:"Frame"`
	Opacity   float64            `editable:"Opacity"`
	BlendFunc concepts.BlendType `editable:"Blend"`
	Tag       string             `editable:"Tag"`

	IgnoreSurfaceTransform bool `editable:"Ignore Surface Transform"`

	Bounded bool             `editable:"Bounded"`
	MinUV   concepts.Vector2 `editable:"Min UV"`
	MaxUV   concepts.Vector2 `editable:"Max UV"`

	mu              sync.Mutex
	lastBoundsFrame uint64
}

func (s *ShaderStage) Construct(data map[string]any) {
	s.Transform = concepts.IdentityMatrix2
	s.Flags = ShaderTiled
	s.Opacity = 1
	s.BlendFunc = concepts.BlendNormal
	s.Tag = ""
	s.MaxUV = concepts.Vector2{1, 1}
	atomic.StoreUint64(&s.lastBoundsFrame, math.MaxUint64)

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
		s.Frame = cast.ToInt(v)
	}

	if v, ok := data["Opacity"]; ok {
		s.Opacity = cast.ToFloat64(v)
	}

	if v, ok := data["BlendingFunc"]; ok {
		s.BlendFunc, _ = concepts.BlendTypeString(cast.ToString(v))
	}

	if v, ok := data["Flags"]; ok {
		s.Flags = concepts.ParseFlags(cast.ToString(v), ShaderFlagsString)
	}

	if v, ok := data["Tag"]; ok {
		s.Tag = cast.ToString(v)
	}

	if v, ok := data["Bounded"]; ok {
		s.Bounded = cast.ToBool(v)
	}
}

func (s *ShaderStage) Serialize() map[string]any {
	result := make(map[string]any)

	if s.Material != 0 {
		result["Material"] = s.Material.Serialize()
	}

	if s.Frame != 0 {
		result["Frame"] = s.Frame
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
	if s.Bounded {
		result["Bounded"] = s.Bounded
	}
	if s.Tag != "" {
		result["Tag"] = s.Tag
	}
	return result
}

// CalculateBounds computes MinUV and MaxUV automatically
// based on the stage Transform, assuming the decal covers the [0, 1] UV space.
func (s *ShaderStage) CalculateBounds() {
	if atomic.LoadUint64(&s.lastBoundsFrame) == ecs.Simulation.Frame {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastBoundsFrame == ecs.Simulation.Frame {
		return
	}

	s.Bounded = true
	det := s.Transform[0]*s.Transform[3] - s.Transform[2]*s.Transform[1]
	if math.Abs(det) < 1e-8 {
		s.MinUV = concepts.Vector2{0, 0}
		s.MaxUV = concepts.Vector2{0, 0}
	} else {
		corners := [4]concepts.Vector2{
			{0, 0},
			{1, 0},
			{0, 1},
			{1, 1},
		}

		s.MinUV[0], s.MinUV[1] = math.MaxFloat64, math.MaxFloat64
		s.MaxUV[0], s.MaxUV[1] = -math.MaxFloat64, -math.MaxFloat64

		for _, c := range corners {
			p := s.Transform.Unproject(&c)
			if p[0] < s.MinUV[0] {
				s.MinUV[0] = p[0]
			}
			if p[0] > s.MaxUV[0] {
				s.MaxUV[0] = p[0]
			}
			if p[1] < s.MinUV[1] {
				s.MinUV[1] = p[1]
			}
			if p[1] > s.MaxUV[1] {
				s.MaxUV[1] = p[1]
			}
		}
	}

	atomic.StoreUint64(&s.lastBoundsFrame, ecs.Simulation.Frame)
}
