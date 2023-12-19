package concepts

import (
	"tlyakhov/gofoom/constants"

	"github.com/loov/hrtime"
)

// This is based on the "Fix your timestep" blog post here:
// https://gafferongames.com/post/fix_your_timestep/
type Simulation struct {
	SimTime          float64
	RenderTime       float64
	FrameMillis      float64
	RenderStateBlend float64
	FPS              float64
	PrevTimestamp    int64
	Timestamp        int64
	Counter          uint64
	Integrate        func()
	Render           func()
	All              map[Simulated]bool
	Animations       map[string]Animated
}

func NewSimulation() *Simulation {
	return &Simulation{
		PrevTimestamp: hrtime.Now().Milliseconds(),
		All:           make(map[Simulated]bool),
		Animations:    make(map[string]Animated),
	}
}

func (s *Simulation) Step() {
	s.Timestamp = hrtime.Now().Milliseconds()
	s.FrameMillis = float64(s.Timestamp - s.PrevTimestamp)
	s.PrevTimestamp = s.Timestamp
	if s.FrameMillis != 0 {
		s.FPS = 1000.0 / s.FrameMillis
	}

	if s.FrameMillis > constants.MinMillisPerFrame {
		s.FrameMillis = constants.MinMillisPerFrame
	}

	s.RenderTime += s.FrameMillis

	for s.RenderTime >= constants.TimeStep {
		for v := range s.All {
			v.NewFrame()
		}

		for _, a := range s.Animations {
			a.Animate()
		}

		if s.Integrate != nil {
			s.Integrate()
		}
		s.Counter++
		s.RenderTime -= constants.TimeStep
		s.SimTime += constants.TimeStep
	}

	// Update the blended values
	s.RenderStateBlend = s.RenderTime / constants.TimeStep

	for v := range s.All {
		v.RenderBlend(s.RenderStateBlend)
	}

	if s.Render != nil {
		s.Render()
	}
}

func (s *Simulation) Animate(name string, a Animated) bool {
	if s.Animations[name] != nil {
		return false
	}
	s.Animations[name] = a
	return true
}
