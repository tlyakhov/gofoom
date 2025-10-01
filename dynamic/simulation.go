// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package dynamic

import (
	"tlyakhov/gofoom/constants"

	"github.com/loov/hrtime"
)

// This is based on the "Fix your timestep" blog post here:
// https://gafferongames.com/post/fix_your_timestep/
type Simulation struct {
	EditorPaused     bool
	Recalculate      bool
	FrameMillis      float64 // Milliseconds
	RenderStateBlend float64
	FPS              float64
	PrevTimestamp    int64 // Milliseconds
	// Wall clock time
	Timestamp int64 // Milliseconds
	// Simulation time (excludes rendering, increments at constant rate)
	SimTimestamp int64 // Milliseconds
	Counter      uint64
	Frame        uint64
	NewFrame     func()
	Integrate    func()
	Render       func()
	Dynamics     map[Dynamic]struct{}   // *xsync.MapOf[Dynamic, struct{}]
	Spawnables   map[Spawnable]struct{} //  *xsync.MapOf[Spawnable, struct{}]
	Events       EventQueue

	simTime    float64
	renderTime float64
}

func NewSimulation() *Simulation {
	return &Simulation{
		PrevTimestamp: hrtime.Now().Milliseconds(),
		Timestamp:     hrtime.Now().Milliseconds(),
		SimTimestamp:  0,
		EditorPaused:  false,
		Dynamics:      make(map[Dynamic]struct{}),
		Spawnables:    make(map[Spawnable]struct{}),
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

	s.renderTime += s.FrameMillis

	if s.NewFrame != nil {
		s.NewFrame()
	}

	for d := range s.Dynamics {
		d.NewFrame()
	}

	for s.renderTime >= constants.TimeStep {
		for d := range s.Dynamics {
			if a := d.GetAnimation(); a != nil && !s.EditorPaused {
				a.Animate()
			}
		}

		if s.Integrate != nil {
			s.Integrate()
		}

		for s.Events.Head != s.Events.Tail {
			s.Events.ConsumeEvent()
		}

		s.Counter++
		s.renderTime -= constants.TimeStep
		s.simTime += constants.TimeStep
	}

	// Update the blended values
	s.RenderStateBlend = s.renderTime / constants.TimeStep

	for d := range s.Dynamics {
		if s.Recalculate {
			d.Recalculate()
		}
		d.Update(s.RenderStateBlend)
	}

	if s.Render != nil {
		s.Render()
		s.Frame++
	}

	/*
		if s.Counter%60 == 0 {
			log.Printf("%v dynamics, %v spawnables", len(s.Dynamics), len(s.Spawnables))
		}
	*/
}

// NewEvent wraps adding a timestamped event to the queue
func (s *Simulation) NewEvent(id EventID, data any) {
	s.Events.PushEvent(&Event{
		ID:        id,
		Timestamp: s.Timestamp,
		Data:      data,
	})
}
