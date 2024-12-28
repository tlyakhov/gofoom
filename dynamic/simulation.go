// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package dynamic

import (
	"tlyakhov/gofoom/constants"

	"github.com/loov/hrtime"
	"github.com/puzpuzpuz/xsync/v3"
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
	Dynamics         *xsync.MapOf[Dynamic, struct{}]
	Spawnables       *xsync.MapOf[Spawnable, struct{}]
}

func NewSimulation() *Simulation {
	return &Simulation{
		PrevTimestamp: hrtime.Now().Milliseconds(),
		Timestamp:     hrtime.Now().Milliseconds(),
		EditorPaused:  false,
		Dynamics:      xsync.NewMapOf[Dynamic, struct{}](),
		Spawnables:    xsync.NewMapOf[Spawnable, struct{}](),
	}
}

func (s *Simulation) Step() {
	s.PrevTimestamp = s.Timestamp
	s.Timestamp = hrtime.Now().Milliseconds()
	s.FrameMillis = float64(s.Timestamp - s.PrevTimestamp)
	if s.FrameMillis != 0 {
		s.FPS = 1000.0 / s.FrameMillis
	}

	if s.FrameMillis > constants.MinMillisPerFrame {
		s.FrameMillis = constants.MinMillisPerFrame
	}

	s.RenderTime += s.FrameMillis

	for s.RenderTime >= constants.TimeStep {
		s.Dynamics.Range(func(d Dynamic, _ struct{}) bool {
			d.NewFrame()
			if a := d.GetAnimation(); a != nil && !s.EditorPaused {
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

	s.Dynamics.Range(func(d Dynamic, _ struct{}) bool {
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
