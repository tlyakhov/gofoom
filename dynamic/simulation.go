// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package dynamic

import (
	"sync"
	"tlyakhov/gofoom/constants"

	"github.com/loov/hrtime"
)

// This is based on the "Fix your timestep" blog post here:
// https://gafferongames.com/post/fix_your_timestep/
type Simulation struct {
	EditorPaused     bool
	Recalculate      bool
	SimTime          float64
	RenderTime       float64
	FrameMillis      float64 // Milliseconds
	RenderStateBlend float64
	FPS              float64
	PrevTimestamp    int64 // Milliseconds
	Timestamp        int64 // Milliseconds
	Counter          uint64
	Frame            uint64
	Integrate        func()
	Render           func()
	All              sync.Map
}

func NewSimulation() *Simulation {
	return &Simulation{
		PrevTimestamp: hrtime.Now().Milliseconds(),
		EditorPaused:  false,
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
		s.All.Range(func(key any, _ any) bool {
			d := key.(Dynamic)
			d.NewFrame()
			if a := d.GetAnimation(); a != nil {
				a.Animate()
			}
			return true
		})

		if s.Integrate != nil {
			s.Integrate()
		}

		s.Counter++
		s.RenderTime -= constants.TimeStep
		s.SimTime += constants.TimeStep
	}

	// Update the blended values
	s.RenderStateBlend = s.RenderTime / constants.TimeStep

	s.All.Range(func(key any, _ any) bool {
		d := key.(Dynamic)
		if s.Recalculate {
			d.Recalculate()
		}
		d.Update(s.RenderStateBlend)
		return true
	})

	if s.Render != nil {
		s.Render()
		s.Frame++
	}
}
