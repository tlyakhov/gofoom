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
	NewFrame         func()
	Integrate        func()
	Render           func()
	Dynamics         *xsync.MapOf[Dynamic, struct{}]
	Spawnables       *xsync.MapOf[Spawnable, struct{}]
	Events           EventQueue
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
	// TODO: We should add some functionality to save all sim values including
	// time to a ledger. Would be great for DOOM style "demos", replays, and
	// also for debugging weird edge cases.
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

	if s.NewFrame != nil {
		s.NewFrame()
	}

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

		for s.Events.Head != s.Events.Tail {
			s.Events.ConsumeEvent()
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

// NewEvent wraps adding a timestamped event to the queue
func (s *Simulation) NewEvent(id EventID, data any) {
	s.Events.PushEvent(&Event{
		ID:        id,
		Timestamp: s.Timestamp,
		Data:      data,
	})
}
