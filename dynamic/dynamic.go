// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package dynamic

import (
	"math"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"

	"github.com/spf13/cast"
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
	Spawned[T] `editable:"^"`
	// The previous frame's value
	Prev T
	// Prior to rendering a frame, this value should be == .Now
	// During rendering, this will be a value blended between .Prev and .Now
	// depending on how much "leftover" dt there is
	// (see https://gafferongames.com/post/fix_your_timestep/)
	Render T

	*Animation[T] `editable:"Animation"`

	IsAngle bool // Only relevant for T=float64

	// Do we need these? not used anywhere currently
	NoRenderBlend bool // Always use next frame value
	OnRender      func(d Dynamic, blend float64)
	OnPostUpdate  func(d Dynamic)

	// Procedural dynamics
	Procedural bool    `editable:"Procedural?"`
	Input      T       `editable:"Input" edit_condition:"IsProcedural"`
	Freq       float64 `editable:"Frequency" edit_condition:"IsProcedural"` // in Hz
	Damping    float64 `editable:"Damping" edit_condition:"IsProcedural"`   // aka Zeta
	Response   float64 `editable:"Response" edit_condition:"IsProcedural"`

	outputV, prevInput T
	k1, k2, k3         float64
}

func (d *DynamicValue[T]) IsProcedural() bool {
	return d.Procedural
}

func (d *DynamicValue[T]) ResetToSpawn() {
	d.Spawned.ResetToSpawn()
	d.Prev = d.Spawn
	d.Render = d.Spawn
	d.Input = d.Spawn
	d.prevInput = d.Spawn
}

func (d *DynamicValue[T]) SetAll(v T) {
	d.Spawn = v
	d.ResetToSpawn()
}

func (d *DynamicValue[T]) Attach(sim *Simulation) {
	sim.Dynamics[d] = struct{}{}
	sim.Spawnables[d] = struct{}{}
	d.Attached = true
}

func (d *DynamicValue[T]) Detach(sim *Simulation) {
	d.Render = d.Now
	d.Attached = false
	delete(sim.Spawnables, d)
	delete(sim.Dynamics, d)
}

func (d *DynamicValue[T]) NewAnimation() *Animation[T] {
	d.Animation = new(Animation[T])
	d.Animation.Construct(nil)
	d.Animation.DynamicValue = d
	return d.Animation
}

func (d *DynamicValue[T]) NewFrame() {
	d.Prev = d.Now
	d.Render = d.Now
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
			concepts.MinimizeAngleDistance(dc.Now, &dc.Input)
			concepts.MinimizeAngleDistance(dc.Input, &dc.prevInput)
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
	case *DynamicValue[concepts.Matrix2]:
		// This seems expensive, probably worth profiling
		// Could at least prevent .[Get|Set]Transform on prevInput, outputV
		prevInputA, prevInputT, _ := dc.prevInput.GetTransform()
		inputA, inputT, _ := dc.Input.GetTransform()
		nowA, nowT, nowS := dc.Now.GetTransform()
		outputVA, outputVT, outputVS := dc.outputV.GetTransform()
		concepts.MinimizeAngleDistance(nowA, &inputA)
		concepts.MinimizeAngleDistance(inputA, &prevInputA)
		inputVA := inputA - prevInputA
		inputVT := concepts.Vector2{
			inputT[0] - prevInputT[0],
			inputT[1] - prevInputT[1],
		}
		/*inputVS := concepts.Vector2{
			inputS[0] / prevInputS[0],
			inputS[1] / prevInputS[1],
		}*/
		nowA += outputVA * dt
		nowT[0] += outputVT[0] * dt
		nowT[1] += outputVT[1] * dt
		//nowS[0] *= outputVS[0] * dt
		//nowS[1] *= outputVS[1] * dt
		dc.Now.SetTransform(nowA, &nowT, &nowS)
		outputVA += dt * (inputA + d.k3*inputVA - nowA - d.k1*outputVA) / k2Stable
		outputVT[0] += dt * (inputT[0] + d.k3*inputVT[0] - nowT[0] - d.k1*outputVT[0]) / k2Stable
		outputVT[1] += dt * (inputT[1] + d.k3*inputVT[1] - nowT[1] - d.k1*outputVT[1]) / k2Stable
		//outputVS[0] *= dt * (inputS[0] * d.k3 * inputVS[0] / nowS[0] / d.k1 * outputVS[0]) / k2Stable
		//outputVS[1] *= dt * (inputS[1] * d.k3 * inputVS[1] / nowS[1] / d.k1 * outputVS[1]) / k2Stable
		/*if outputVS[0] == 0 {
			outputVS[0] = 1
		}
		if outputVS[1] == 0 {
			outputVS[1] = 1
		}*/
		dc.outputV.SetTransform(outputVA, &outputVT, &outputVS)
	default:
		panic("DynamicValue[T] procedural animations only implemented for float64, vector2, vector3, matrix2")
	}
	d.prevInput = d.Input
}
func (d *DynamicValue[T]) Update(blend float64) {
	if d.Procedural {
		d.UpdateProcedural()
	}
	if d.OnPostUpdate != nil {
		d.OnPostUpdate(d)
	}
	if d.OnRender != nil {
		defer d.OnRender(d, blend)
	}

	// The check for prev == now lets us avoid the linear interpolation, which
	// can introduce precision errors for static float quantities.
	if d.NoRenderBlend || d.Prev == d.Now {
		d.Render = d.Now
		return
	}
	switch dc := any(d).(type) {
	case *DynamicValue[int]:
		dc.Render = int(Lerp(float64(dc.Prev), float64(dc.Now), blend))
	case *DynamicValue[float64]:
		if dc.IsAngle {
			dc.Render = TweenAngles(dc.Prev, dc.Now, blend, Lerp)
		} else {
			dc.Render = Lerp(dc.Prev, dc.Now, blend)
		}
	case *DynamicValue[concepts.Vector2]:
		dc.Render[0] = Lerp(dc.Prev[0], dc.Now[0], blend)
		dc.Render[1] = Lerp(dc.Prev[1], dc.Now[1], blend)
	case *DynamicValue[concepts.Vector3]:
		dc.Render[0] = Lerp(dc.Prev[0], dc.Now[0], blend)
		dc.Render[1] = Lerp(dc.Prev[1], dc.Now[1], blend)
		dc.Render[2] = Lerp(dc.Prev[2], dc.Now[2], blend)
	case *DynamicValue[concepts.Vector4]:
		dc.Render[0] = Lerp(dc.Prev[0], dc.Now[0], blend)
		dc.Render[1] = Lerp(dc.Prev[1], dc.Now[1], blend)
		dc.Render[2] = Lerp(dc.Prev[2], dc.Now[2], blend)
		dc.Render[3] = Lerp(dc.Prev[3], dc.Now[3], blend)
	case *DynamicValue[concepts.Matrix2]:
		d.Render = d.Now
	}
}

func (d *DynamicValue[T]) Serialize() any {
	result := make(map[string]any)

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

	if len(result) > 0 {
		result["Spawn"] = d.serializeValue(d.Spawn)
		result["Now"] = d.serializeValue(d.Now)
		return result
	} else {
		return d.Spawned.Serialize()
	}
}

func (d *DynamicValue[T]) Construct(data any) {
	d.Spawned.Construct(data)
	d.Freq = 4.58
	d.Damping = 0.35
	d.Response = -3.54
	d.Prev = d.Now
	d.Input = d.Now
	d.prevInput = d.Now
	// Ensure we have a reasonable value for render prior to simulation update
	d.Render = d.Now

	if data == nil {
		return
	}
	var params map[string]any
	var ok bool
	if params, ok = data.(map[string]any); !ok || params["Spawn"] == nil {
		return
	}

	if v, ok := params["Procedural"]; ok {
		d.Procedural = cast.ToBool(v)
	}
	if v, ok := params["Freq"]; ok {
		d.Freq = cast.ToFloat64(v)
	}
	if v, ok := params["Damping"]; ok {
		d.Damping = cast.ToFloat64(v)
	}
	if v, ok := params["Response"]; ok {
		d.Response = cast.ToFloat64(v)
	}
	//	TODO: Serialize Input as well?

	d.Recalculate()

	if v, ok := params["Animation"]; ok {
		d.Animation = new(Animation[T])
		d.Animation.Construct(v.(map[string]any))
		d.Animation.DynamicValue = d
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

	switch dc := any(d).(type) {
	case *DynamicValue[concepts.Matrix2]:
		dc.outputV.SetIdentity()
	}
}
