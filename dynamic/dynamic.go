// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package dynamic

import (
	"log"
	"math"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"

	"github.com/spf13/cast"
)

//go:generate go run github.com/dmarkham/enumer -type=DynamicStage -json
type DynamicStage int

const (
	DynamicSpawn DynamicStage = iota
	DynamicPrev
	DynamicRender
	DynamicNow
)

// A DynamicValue is anything in the engine that evolves over time.
//
// * They store several states:
//   - Spawn (what to use when loading a world or respawning)
//   - Prev/Now (previous frame, next frame)
//   - Render (a value blended between Prev & Now based on last frame time)
//   - Input (if this is a procedurally animated value, this is the input)
//
// * Procedurally animated values use second-order dynamics for organic movement
// * Values can also have Animations that use easing (e.g. inventory item bobbing)
type DynamicValue[T DynamicType] struct {
	*Animation[T] `editable:"Animation"`
	Spawned[T]    `editable:"^"`

	Prev T
	// If there are runtime errors about this field being nil, it's probably
	// because the .Attach() method was never called
	Render  *T
	IsAngle bool // Only relevant for T=float64

	// Do we need these? not used anywhere currently
	NoRenderBlend bool // Always use next frame value
	OnRender      func(blend float64)

	// Procedural dynamics
	Procedural bool    `editable:"Procedural?"`
	Input      T       `editable:"Input" edit_condition:"IsProcedural"`
	Freq       float64 `editable:"Frequency" edit_condition:"IsProcedural"` // in Hz
	Damping    float64 `editable:"Damping" edit_condition:"IsProcedural"`   // aka Zeta
	Response   float64 `editable:"Response" edit_condition:"IsProcedural"`

	outputV, prevInput T
	k1, k2, k3         float64

	render T
}

func (d *DynamicValue[T]) IsProcedural() bool {
	return d.Procedural
}

func (d *DynamicValue[T]) Value(s DynamicStage) T {
	switch s {
	case DynamicSpawn:
		return d.Spawn
	case DynamicPrev:
		return d.Prev
	case DynamicNow:
		return d.Now
	default:
		return *d.Render
	}
}

func (d *DynamicValue[T]) ResetToSpawn() {
	if d.Render == nil {
		log.Println("DynamicValue[T].ResetToSpawn: value is unattached to Simulation. Stack trace:")
		log.Println(concepts.StackTrace())
		return
	}
	d.Spawned.ResetToSpawn()
	d.Prev = d.Spawn
	*d.Render = d.Spawn
	d.Input = d.Spawn
	d.prevInput = d.Spawn
}

func (d *DynamicValue[T]) SetAll(v T) {
	d.Spawn = v
	d.ResetToSpawn()
}

func (d *DynamicValue[T]) Attach(sim *Simulation) {
	sim.Dynamics.Store(d, struct{}{})
	sim.Spawnables.Store(d, struct{}{})
	d.Attached = true
	d.Render = &d.render
}

func (d *DynamicValue[T]) Detach(sim *Simulation) {
	d.Render = &d.Now
	d.Attached = false
	sim.Spawnables.Delete(d)
	sim.Dynamics.Delete(d)
}

func (d *DynamicValue[T]) NewAnimation() *Animation[T] {
	d.Animation = new(Animation[T])
	d.Animation.Construct(nil)
	d.Animation.DynamicValue = d
	return d.Animation
}

func (d *DynamicValue[T]) NewFrame() {
	d.Prev = d.Now
}

func fixAngle(src float64, dst *float64) {
	for *dst-src > 180 {
		*dst -= 360
	}
	for *dst-src < -180 {
		*dst += 360
	}
}

func (d *DynamicValue[T]) UpdateProcedural() {
	// Based on "Giving Personality to Procedural Animations using Math"
	// https://www.youtube.com/watch?v=KPoeNZZ6H4s
	dt := constants.TimeStepS
	k2Stable := math.Max(math.Max(d.k2, dt*dt*0.5+dt*d.k1*0.5), dt*d.k1)
	switch dc := any(d).(type) {
	case *DynamicValue[float64]:
		if dc.IsAngle {
			// |Input-prevInput| should be < 180
			// |Input-Now| should be < 180
			fixAngle(dc.Now, &dc.Input)
			fixAngle(dc.Input, &dc.prevInput)
		}
		inputV := dc.Input - dc.prevInput
		dc.Now += dc.outputV * dt
		dc.outputV += dt * (dc.Input + d.k3*inputV - dc.Now - d.k1*dc.outputV) / k2Stable
	case *DynamicValue[concepts.Vector2]:
		inputV := concepts.Vector2{
			dc.Input[0] - dc.prevInput[0],
			dc.Input[1] - dc.prevInput[1]}
		dc.Now[0] += dc.outputV[0] * dt
		dc.Now[1] += dc.outputV[1] * dt
		dc.outputV[1] += dt * (dc.Input[1] + d.k3*inputV[1] - dc.Now[1] - d.k1*dc.outputV[1]) / k2Stable
		dc.outputV[0] += dt * (dc.Input[0] + d.k3*inputV[0] - dc.Now[0] - d.k1*dc.outputV[0]) / k2Stable
	case *DynamicValue[concepts.Vector3]:
		inputV := concepts.Vector3{
			dc.Input[0] - dc.prevInput[0],
			dc.Input[1] - dc.prevInput[1],
			dc.Input[2] - dc.prevInput[2]}
		dc.Now[0] += dc.outputV[0] * dt
		dc.Now[1] += dc.outputV[1] * dt
		dc.Now[2] += dc.outputV[2] * dt
		dc.outputV[2] += dt * (dc.Input[2] + d.k3*inputV[2] - dc.Now[2] - d.k1*dc.outputV[2]) / k2Stable
		dc.outputV[1] += dt * (dc.Input[1] + d.k3*inputV[1] - dc.Now[1] - d.k1*dc.outputV[1]) / k2Stable
		dc.outputV[0] += dt * (dc.Input[0] + d.k3*inputV[0] - dc.Now[0] - d.k1*dc.outputV[0]) / k2Stable
	default:
		panic("DynamicValue[T] procedural animations only implemented for float64, vector2, vector3")
	}
	d.prevInput = d.Input
}
func (d *DynamicValue[T]) Update(blend float64) {
	if d.Procedural {
		d.UpdateProcedural()
	}
	if d.NoRenderBlend {
		d.Render = &d.Now
		if d.OnRender != nil {
			d.OnRender(blend)
		}
		return
	}
	switch dc := any(d).(type) {
	case *DynamicValue[int]:
		dc.render = int(Lerp(float64(dc.Prev), float64(dc.Now), blend))
	case *DynamicValue[float64]:
		if dc.IsAngle {
			dc.render = TweenAngles(dc.Prev, dc.Now, blend, Lerp)
		} else {
			dc.render = Lerp(dc.Prev, dc.Now, blend)
		}
	case *DynamicValue[concepts.Vector2]:
		dc.render[0] = Lerp(dc.Prev[0], dc.Now[0], blend)
		dc.render[1] = Lerp(dc.Prev[1], dc.Now[1], blend)
	case *DynamicValue[concepts.Vector3]:
		dc.render[0] = Lerp(dc.Prev[0], dc.Now[0], blend)
		dc.render[1] = Lerp(dc.Prev[1], dc.Now[1], blend)
		dc.render[2] = Lerp(dc.Prev[2], dc.Now[2], blend)
	case *DynamicValue[concepts.Vector4]:
		dc.render[0] = Lerp(dc.Prev[0], dc.Now[0], blend)
		dc.render[1] = Lerp(dc.Prev[1], dc.Now[1], blend)
		dc.render[2] = Lerp(dc.Prev[2], dc.Now[2], blend)
		dc.render[3] = Lerp(dc.Prev[3], dc.Now[3], blend)
	case *DynamicValue[concepts.Matrix2]:
		d.Render = &d.Now
	}

	if d.OnRender != nil {
		d.OnRender(blend)
	}
}

func (d *DynamicValue[T]) Serialize() map[string]any {
	result := d.Spawned.Serialize()

	if d.Procedural {
		result["Procedural"] = d.Procedural
	}
	if d.Freq != 4.58 {
		result["Freq"] = d.Freq
	}
	if d.Damping != 0.35 {
		result["Damping"] = d.Damping
	}
	if d.Response != -3.54 {
		result["Response"] = d.Response
	}

	if d.Animation != nil {
		result["Animation"] = d.Animation.Serialize()
	}
	return result
}

func (d *DynamicValue[T]) Construct(data map[string]any) {
	d.Spawned.Construct(data)
	d.Freq = 4.58
	d.Damping = 0.35
	d.Response = -3.54
	d.Prev = d.Now
	d.Input = d.Now
	d.prevInput = d.Now
	if !d.Attached {
		d.Render = &d.Now
	}

	if data == nil {
		return
	}

	if v, ok := data["Procedural"]; ok {
		d.Procedural = cast.ToBool(v)
	}
	if v, ok := data["Freq"]; ok {
		d.Freq = cast.ToFloat64(v)
	}
	if v, ok := data["Damping"]; ok {
		d.Damping = cast.ToFloat64(v)
	}
	if v, ok := data["Response"]; ok {
		d.Response = cast.ToFloat64(v)
	}
	//	TODO: Serialize Input as well?

	d.Recalculate()

	if v, ok := data["Animation"]; ok {
		d.Animation = new(Animation[T])
		d.Animation.Construct(v.(map[string]any))
		d.Animation.DynamicValue = d
	}

	// Ensure we have a reasonable value for render prior to simulation update
	if d.Attached {
		d.render = d.Now
	}
}

func (d *DynamicValue[T]) GetAnimation() Animated {
	return d.Animation
}

func (d *DynamicValue[T]) Recalculate() {
	// Based on "Giving Personality to Procedural Animations using Math"
	// https://www.youtube.com/watch?v=KPoeNZZ6H4s
	if d.Freq == 0 {
		d.Freq = 0.000001
	}
	radians := (2.0 * math.Pi * d.Freq)
	d.k1 = d.Damping / (math.Pi * d.Freq)
	d.k2 = 1.0 / (radians * radians)
	d.k3 = d.Response * d.Damping / radians
	// 80% of actual limit, to be safe
	// d.tCrit = 0.8 * (math.Sqrt(4*d.k2+d.k1*d.k1) - d.k1)
	var zero T
	d.outputV = zero
}
