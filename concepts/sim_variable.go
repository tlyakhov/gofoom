package concepts

import "log"

type Simulatable interface {
	~int | ~float64 | Vector2 | Vector3 | Vector4
}

type Simulated interface {
	Attach(sim *Simulation)
	Detach(sim *Simulation)
	Reset()
	Blend(float64)
	NewFrame()
}

type SimVariable[T Simulatable] struct {
	Now            T
	Prev           T
	Original       T `editable:"Initial Value"`
	Render         T
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

func (s *SimVariable[T]) NewFrame() {
	s.Prev = s.Now
}

func (s *SimVariable[T]) Blend(blend float64) {
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

func (s *SimVariable[T]) Serialize() any {
	switch sc := any(s).(type) {
	case *SimVariable[int]:
		return sc.Original
	case *SimVariable[float64]:
		return sc.Original
	case *SimVariable[Vector2]:
		return sc.Original.Serialize()
	case *SimVariable[Vector3]:
		return sc.Original.Serialize()
	case *SimVariable[Vector4]:
		return sc.Original.Serialize(false)
	default:
		log.Panicf("Tried to serialize SimVar[T] %v where T has no serializer", s)
	}
	return nil
}

func (s *SimVariable[T]) Deserialize(data any) {
	switch sc := any(s).(type) {
	case *SimVariable[int]:
		sc.Original = data.(int)
	case *SimVariable[float64]:
		sc.Original = data.(float64)
	case *SimVariable[Vector2]:
		sc.Original.Deserialize(data.(map[string]any))
	case *SimVariable[Vector3]:
		sc.Original.Deserialize(data.(map[string]any))
	case *SimVariable[Vector4]:
		sc.Original.Deserialize(data.(map[string]any), false)
	default:
		log.Panicf("Tried to deserialize SimVar[T] %v where T has no serializer", s)
	}
	s.Reset()
}

// For scripting
func (s *SimVariable[T]) Ptr() *SimVariable[T] {
	return s
}
