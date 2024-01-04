package concepts

import (
	"log"
	"reflect"
	"regexp"
	"strconv"
)

type Simulatable interface {
	~int | ~float64 | Vector2 | Vector3 | Vector4
}

type Simulated interface {
	Serializable
	Attach(sim *Simulation)
	Detach(sim *Simulation)
	Reset()
	RenderBlend(float64)
	NewFrame()
	GetAnimation() Animated
}

type SimVariable[T Simulatable] struct {
	*Animation[T] `editable:"Animation"`

	Now            T
	Prev           T
	Original       T `editable:"Initial Value"`
	Render         T
	NoRenderBlend  bool // For things like angles
	RenderCallback func()
}

func (s *SimVariable[T]) Reset() {
	s.Prev = s.Original
	s.Now = s.Original
}

func (s *SimVariable[T]) Set(v T) {
	s.Original = v
	s.Reset()
}

func (s *SimVariable[T]) Attach(sim *Simulation) {
	sim.All[s] = true
}

func (s *SimVariable[T]) Detach(sim *Simulation) {
	delete(sim.All, s)
}

func (s *SimVariable[T]) NewAnimation() *Animation[T] {
	s.Animation = new(Animation[T])
	s.Animation.Construct(nil)
	s.Animation.SimVariable = s
	return s.Animation
}

func (s *SimVariable[T]) NewFrame() {
	s.Prev = s.Now
}

func (s *SimVariable[T]) RenderBlend(blend float64) {
	if s.NoRenderBlend {
		s.Render = s.Now
		return
	}
	switch sc := any(s).(type) {
	case *SimVariable[int]:
		sc.Render = int(Lerp(float64(sc.Prev), float64(sc.Now), blend))
	case *SimVariable[float64]:
		sc.Render = Lerp(sc.Prev, sc.Now, blend)
	case *SimVariable[Vector2]:
		sc.Render[0] = Lerp(sc.Prev[0], sc.Now[0], blend)
		sc.Render[1] = Lerp(sc.Prev[1], sc.Now[1], blend)
	case *SimVariable[Vector3]:
		sc.Render[0] = Lerp(sc.Prev[0], sc.Now[0], blend)
		sc.Render[1] = Lerp(sc.Prev[1], sc.Now[1], blend)
		sc.Render[2] = Lerp(sc.Prev[2], sc.Now[2], blend)
	case *SimVariable[Vector4]:
		sc.Render[0] = Lerp(sc.Prev[0], sc.Now[0], blend)
		sc.Render[1] = Lerp(sc.Prev[1], sc.Now[1], blend)
		sc.Render[2] = Lerp(sc.Prev[2], sc.Now[2], blend)
		sc.Render[3] = Lerp(sc.Prev[3], sc.Now[3], blend)
	}

	if s.RenderCallback != nil {
		s.RenderCallback()
	}
}

func (s *SimVariable[T]) Serialize() map[string]any {
	result := make(map[string]any)

	switch sc := any(s).(type) {
	case *SimVariable[int]:
		result["Original"] = sc.Original
	case *SimVariable[float64]:
		result["Original"] = sc.Original
	case *SimVariable[Vector2]:
		result["Original"] = sc.Original.Serialize()
	case *SimVariable[Vector3]:
		result["Original"] = sc.Original.Serialize()
	case *SimVariable[Vector4]:
		result["Original"] = sc.Original.Serialize(false)
	default:
		log.Panicf("Tried to serialize SimVar[T] %v where T has no serializer", s)
	}

	if s.Animation != nil {
		result["Animation"] = s.Animation.Serialize()
	}
	return result
}

func (s *SimVariable[T]) Construct(data map[string]any) {
	if data == nil {
		return
	}

	if v, ok := data["Original"]; ok {
		switch sc := any(s).(type) {
		case *SimVariable[int]:
			sc.Original = v.(int)
		case *SimVariable[float64]:
			sc.Original = v.(float64)
		case *SimVariable[Vector2]:
			sc.Original.Deserialize(v.(map[string]any))
		case *SimVariable[Vector3]:
			sc.Original.Deserialize(v.(map[string]any))
		case *SimVariable[Vector4]:
			sc.Original.Deserialize(v.(map[string]any), false)
		default:
			log.Panicf("Tried to deserialize SimVar[T] %v where T has no serializer", s)
		}
		s.Reset()
	}

	if v, ok := data["Animation"]; ok {
		s.Animation = new(Animation[T])
		s.Animation.Construct(v.(map[string]any))
		s.Animation.SimVariable = s
	}
}

func (s *SimVariable[T]) GetAnimation() Animated {
	return s.Animation
}

// entity.component.field (e.g. "53.Body.Pos")
var reSimVariableSource = regexp.MustCompile(`(\d+).(\w+).(\w+)`)

func SimVariableFromString[T Simulatable](db *EntityComponentDB, source string) *SimVariable[T] {
	matches := reSimVariableSource.FindStringSubmatch(source)
	if len(matches) != 3 {
		log.Printf("SimVariableFromString - target %v isn't entity.component.field", source)
		return nil
	}
	entity, _ := strconv.ParseUint(matches[0], 10, 64)
	index := DbTypes().Indexes[matches[1]]
	field := matches[2]
	component := db.EntityComponents[entity][index]
	if component == nil {
		log.Printf("SimVariableFromString - component %v on entity %v is nil", matches[1], entity)
		return nil
	}
	fieldValue := reflect.ValueOf(component).Elem().FieldByName(field)
	if fieldValue.IsZero() {
		log.Printf("SimVariableFromString - no field %v on component %v", field, component.String())
		return nil
	}
	if result, ok := fieldValue.Addr().Interface().(*SimVariable[T]); ok {
		return result
	} else {
		log.Printf("SimVariableFromString - %v is not *SimVariable[T]", source)
		return nil
	}
}
