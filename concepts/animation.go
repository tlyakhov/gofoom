package concepts

import (
	"tlyakhov/gofoom/constants"

	"github.com/rs/xid"
)

//go:generate go run github.com/dmarkham/enumer -type=AnimationStyle -json
type AnimationStyle int

const (
	AnimationStyleOnce AnimationStyle = iota
	AnimationStyleHold
	AnimationStyleLoop
)

type IAnimation interface {
	Serializable
	Animate()
	GetName() string
}

type Animation[T Simulatable] struct {
	*Simulation
	TweeningFunc `editable:"Tweening Function"`

	Name     string          `editable:"Name"`
	Start    T               `editable:"Start"`
	End      T               `editable:"End"`
	Duration float64         `editable:"Duration"`
	Active   bool            `editable:"Active"`
	Reverse  bool            `editable:"Reverse"`
	Style    AnimationStyle  `editable:"Style"`
	Target   *SimVariable[T] `editable:"Target"`
	Percent  float64
}

func (a *Animation[T]) SetEasingFunc(name string) TweeningFunc {
	a.TweeningFunc = TweeningFuncs[name]
	if a.TweeningFunc == nil {
		a.TweeningFunc = Lerp
	}
	return a.TweeningFunc
}

func (a *Animation[T]) Animate() {
	if !a.Active || a.Target == nil {
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

	if (a.Percent >= 1 && !a.Reverse) || (a.Percent <= 0 && a.Reverse) {
		switch a.Style {
		case AnimationStyleOnce:
			delete(a.Animations, a.Name)
		case AnimationStyleLoop:
			a.Reverse = !a.Reverse
		}
	}
	if a.Reverse {
		a.Percent -= constants.TimeStep / a.Duration
	} else {
		a.Percent += constants.TimeStep / a.Duration
	}
}

func (a *Animation[T]) Construct(data map[string]any) {
	a.Active = true
	a.Duration = 1000
	a.Reverse = false
	a.TweeningFunc = Lerp
	a.Style = AnimationStyleLoop
	a.Name = xid.New().String()

	if data == nil {
		return
	}

	// TODO: Construction
}
func (a *Animation[T]) Serialize() map[string]any {
	result := make(map[string]any)
	// TODO: Serialization
	return result
}

func (a *Animation[T]) SetDB(db *EntityComponentDB) {
	a.Simulation = db.Simulation
}

func (a *Animation[T]) GetName() string {
	return a.Name
}
