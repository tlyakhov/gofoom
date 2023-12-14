package concepts

import (
	"tlyakhov/gofoom/constants"
)

//go:generate go run github.com/dmarkham/enumer -type=AnimationStyle -json
type AnimationStyle int

const (
	AnimationStyleOnce AnimationStyle = iota
	AnimationStyleHold
)

type Animated interface {
	Animate()
}

type Animation[T Simulatable] struct {
	*Simulation
	EasingFunc
	Name       string
	Target     *SimVariable[T]
	Start, End T
	Duration   float64
	Active     bool
	Reverse    bool
	Percent    float64
	Style      AnimationStyle
}

func NewAnimation[T Simulatable](s *Simulation, name string, target *SimVariable[T], start, end T, duration float64) *Animation[T] {
	a := Animation[T]{
		Active:     true,
		Reverse:    false,
		Simulation: s,
		Name:       name,
		Target:     target,
		Start:      start,
		End:        end,
		Duration:   duration,
		EasingFunc: Lerp,
		Style:      AnimationStyleOnce,
	}
	return &a
}

func (a *Animation[T]) SetEasingFunc(name string) EasingFunc {
	a.EasingFunc = EasingFuncs[name]
	if a.EasingFunc == nil {
		a.EasingFunc = Lerp
	}
	return a.EasingFunc
}

func (a *Animation[T]) Animate() {
	if !a.Active {
		return
	}
	a.Percent = Clamp(a.Percent, 0, 1)
	switch c := any(a).(type) {
	case *Animation[int]:
		c.Target.Now = int(c.EasingFunc(float64(c.Start), float64(c.End), c.Percent))
	case *Animation[float64]:
		c.Target.Now = c.EasingFunc(c.Start, c.End, c.Percent)
	case *Animation[Vector2]:
		c.Target.Now[0] = c.EasingFunc(c.Start[0], c.End[0], c.Percent)
		c.Target.Now[1] = c.EasingFunc(c.Start[1], c.End[1], c.Percent)
	case *Animation[Vector3]:
		c.Target.Now[0] = c.EasingFunc(c.Start[0], c.End[0], c.Percent)
		c.Target.Now[1] = c.EasingFunc(c.Start[1], c.End[1], c.Percent)
		c.Target.Now[2] = c.EasingFunc(c.Start[2], c.End[2], c.Percent)
	case *Animation[Vector4]:
		c.Target.Now[0] = c.EasingFunc(c.Start[0], c.End[0], c.Percent)
		c.Target.Now[1] = c.EasingFunc(c.Start[1], c.End[1], c.Percent)
		c.Target.Now[2] = c.EasingFunc(c.Start[2], c.End[2], c.Percent)
		c.Target.Now[3] = c.EasingFunc(c.Start[3], c.End[3], c.Percent)
	}

	if a.Percent >= 1 && a.Style == AnimationStyleOnce {
		delete(a.Animations, a.Name)
	}
	if a.Reverse {
		a.Percent -= constants.TimeStep / a.Duration
	} else {
		a.Percent += constants.TimeStep / a.Duration
	}
}
