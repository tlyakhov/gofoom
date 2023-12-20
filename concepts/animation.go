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
	TweeningFunc
	Name       string
	Target     *SimVariable[T]
	Start, End T
	Duration   float64
	Active     bool
	Reverse    bool
	Percent    float64
	Style      AnimationStyle
}

func (a *Animation[T]) SetEasingFunc(name string) TweeningFunc {
	a.TweeningFunc = TweeningFuncs[name]
	if a.TweeningFunc == nil {
		a.TweeningFunc = Lerp
	}
	return a.TweeningFunc
}

func (a *Animation[T]) Animate() {
	if !a.Active {
		return
	}
	a.Percent = Clamp(a.Percent, 0, 1)
	switch c := any(a).(type) {
	case *Animation[int]:
		c.Target.Now = int(c.TweeningFunc(float64(c.Start), float64(c.End), c.Percent))
	case *Animation[float64]:
		c.Target.Now = c.TweeningFunc(c.Start, c.End, c.Percent)
	case *Animation[Vector2]:
		c.Target.Now[0] = c.TweeningFunc(c.Start[0], c.End[0], c.Percent)
		c.Target.Now[1] = c.TweeningFunc(c.Start[1], c.End[1], c.Percent)
	case *Animation[Vector3]:
		c.Target.Now[0] = c.TweeningFunc(c.Start[0], c.End[0], c.Percent)
		c.Target.Now[1] = c.TweeningFunc(c.Start[1], c.End[1], c.Percent)
		c.Target.Now[2] = c.TweeningFunc(c.Start[2], c.End[2], c.Percent)
	case *Animation[Vector4]:
		c.Target.Now[0] = c.TweeningFunc(c.Start[0], c.End[0], c.Percent)
		c.Target.Now[1] = c.TweeningFunc(c.Start[1], c.End[1], c.Percent)
		c.Target.Now[2] = c.TweeningFunc(c.Start[2], c.End[2], c.Percent)
		c.Target.Now[3] = c.TweeningFunc(c.Start[3], c.End[3], c.Percent)
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

// TODO: Add serialization?
func (a *Animation[T]) Construct(s *Simulation) {
	a.Simulation = s
	a.Active = true
	a.Reverse = false
	a.TweeningFunc = Lerp
	a.Style = AnimationStyleOnce
}
