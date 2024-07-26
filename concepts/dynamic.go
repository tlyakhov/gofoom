// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"log"
	"reflect"
	"regexp"
	"strconv"
)

// Dynamic is an interface for any value that is affected by time in the engine:
// 1. They have a lifecycle with a starting value that changes over time
// 2. They may have a "render" value interpolated between a past/future values.
type Dynamic interface {
	Serializable
	Attach(sim *Simulation)
	Detach(sim *Simulation)
	ResetToOriginal()
	RenderBlend(float64)
	NewFrame()
	GetAnimation() Animated
}

//go:generate go run github.com/dmarkham/enumer -type=DynamicStage -json
type DynamicStage int

const (
	DynamicOriginal DynamicStage = iota
	DynamicPrev
	DynamicRender
	DynamicNow
)

// A DynamicType is a type constraint for anything the engine can simulate
type DynamicType interface {
	~int | ~float64 | Vector2 | Vector3 | Vector4 | Matrix2
}

type DynamicValue[T DynamicType] struct {
	*Animation[T] `editable:"Animation"`

	Now            T
	Prev           T
	Original       T `editable:"Initial Value"`
	Render         *T
	Attached       bool
	NoRenderBlend  bool // For things like angles
	RenderCallback func(blend float64)

	render T
}

func (d *DynamicValue[T]) ResetToOriginal() {
	d.Prev = d.Original
	d.Now = d.Original
}

func (d *DynamicValue[T]) SetAll(v T) {
	d.Original = v
	d.ResetToOriginal()
}

func (d *DynamicValue[T]) Attach(sim *Simulation) {
	sim.All.Store(d, true)
	d.Render = &d.render
	d.Attached = true
}

func (d *DynamicValue[T]) Detach(sim *Simulation) {
	sim.All.Delete(d)
	d.Render = &d.Now
	d.Attached = false
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

func (d *DynamicValue[T]) RenderBlend(blend float64) {
	if d.NoRenderBlend {
		d.Render = &d.Now
		if d.RenderCallback != nil {
			d.RenderCallback(blend)
		}
		return
	}
	switch dc := any(d).(type) {
	case *DynamicValue[int]:
		dc.render = int(Lerp(float64(dc.Prev), float64(dc.Now), blend))
	case *DynamicValue[float64]:
		dc.render = Lerp(dc.Prev, dc.Now, blend)
	case *DynamicValue[Vector2]:
		dc.render[0] = Lerp(dc.Prev[0], dc.Now[0], blend)
		dc.render[1] = Lerp(dc.Prev[1], dc.Now[1], blend)
	case *DynamicValue[Vector3]:
		dc.render[0] = Lerp(dc.Prev[0], dc.Now[0], blend)
		dc.render[1] = Lerp(dc.Prev[1], dc.Now[1], blend)
		dc.render[2] = Lerp(dc.Prev[2], dc.Now[2], blend)
	case *DynamicValue[Vector4]:
		dc.render[0] = Lerp(dc.Prev[0], dc.Now[0], blend)
		dc.render[1] = Lerp(dc.Prev[1], dc.Now[1], blend)
		dc.render[2] = Lerp(dc.Prev[2], dc.Now[2], blend)
		dc.render[3] = Lerp(dc.Prev[3], dc.Now[3], blend)
	case *DynamicValue[Entity]:
		dc.render = dc.Prev
		if blend > 0.5 {
			dc.render = dc.Now
		}
	case *DynamicValue[Matrix2]:
		d.Render = &d.Now
	}

	if d.RenderCallback != nil {
		d.RenderCallback(blend)
	}
}

func (d *DynamicValue[T]) Serialize() map[string]any {
	result := make(map[string]any)

	switch dc := any(d).(type) {
	case *DynamicValue[int]:
		result["Original"] = dc.Original
	case *DynamicValue[float64]:
		result["Original"] = dc.Original
	case *DynamicValue[Vector2]:
		result["Original"] = dc.Original.Serialize()
	case *DynamicValue[Vector3]:
		result["Original"] = dc.Original.Serialize()
	case *DynamicValue[Vector4]:
		result["Original"] = dc.Original.Serialize(false)
	case *DynamicValue[Matrix2]:
		result["Original"] = dc.Original.Serialize()
	case *DynamicValue[Entity]:
		result["Original"] = dc.Original.Format()
	default:
		log.Panicf("Tried to serialize SimVar[T] %v where T has no serializer", d)
	}

	if d.Animation != nil {
		result["Animation"] = d.Animation.Serialize()
	}
	return result
}

func (d *DynamicValue[T]) Construct(data map[string]any) {
	if !d.Attached {
		d.Render = &d.Now
	}

	switch sc := any(d).(type) {
	case *DynamicValue[Matrix2]:
		sc.Original.SetIdentity()
	}

	if data == nil {
		d.ResetToOriginal()
		return
	}

	if v, ok := data["Original"]; ok {
		switch dc := any(d).(type) {
		case *DynamicValue[int]:
			dc.Original = v.(int)
		case *DynamicValue[float64]:
			dc.Original = v.(float64)
		case *DynamicValue[Vector2]:
			dc.Original.Deserialize(v.(map[string]any))
		case *DynamicValue[Vector3]:
			dc.Original.Deserialize(v.(map[string]any))
		case *DynamicValue[Vector4]:
			dc.Original.Deserialize(v.(map[string]any), false)
		case *DynamicValue[Matrix2]:
			dc.Original.Deserialize(v.([]any))
		default:
			log.Panicf("Tried to deserialize SimVar[T] %v where T has no serializer", d)
		}
	}

	d.ResetToOriginal()

	if v, ok := data["Animation"]; ok {
		d.Animation = new(Animation[T])
		d.Animation.Construct(v.(map[string]any))
		d.Animation.DynamicValue = d
	}
}

func (d *DynamicValue[T]) GetAnimation() Animated {
	return d.Animation
}

// entity.component.field (e.g. "53.Body.Pos")
var reDynamicSource = regexp.MustCompile(`(\d+).(\w+).(\w+)`)

func DynamicFromString[T DynamicType](db *EntityComponentDB, source string) *DynamicValue[T] {
	matches := reDynamicSource.FindStringSubmatch(source)
	if len(matches) != 3 {
		log.Printf("DynamicFromString - target %v isn't entity.component.field", source)
		return nil
	}
	entity, _ := strconv.ParseUint(matches[0], 10, 64)
	index := DbTypes().Indexes[matches[1]]
	field := matches[2]
	component := db.EntityComponents[entity][index]
	if component == nil {
		log.Printf("DynamicFromString - component %v on entity %v is nil", matches[1], entity)
		return nil
	}
	fieldValue := reflect.ValueOf(component).Elem().FieldByName(field)
	if fieldValue.IsZero() {
		log.Printf("DynamicFromString - no field %v on component %v", field, component.String())
		return nil
	}
	if result, ok := fieldValue.Addr().Interface().(*DynamicValue[T]); ok {
		return result
	} else {
		log.Printf("DynamicFromString - %v is not *DynamicValue[T]", source)
		return nil
	}
}
