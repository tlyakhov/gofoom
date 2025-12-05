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
	FrameNanos       int64 // Nanoseconds
	RenderStateBlend float64
	FPS              float64
	PrevTimestamp    int64 // Nanoseconds
	// Wall clock time
	Timestamp int64 // Nanoseconds
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

	renderTime int64
}

func NewSimulation() *Simulation {
	return &Simulation{
		PrevTimestamp: hrtime.Now().Nanoseconds(),
		Timestamp:     hrtime.Now().Nanoseconds(),
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
	s.Timestamp = hrtime.Now().Nanoseconds()
	s.FrameNanos = s.Timestamp - s.PrevTimestamp
	if s.FrameNanos != 0 {
		s.FPS = float64(1_000_000_000) / float64(s.FrameNanos)
	}

	if s.FrameNanos > constants.MinMillisPerFrame*1_000_000 {
		s.FrameNanos = constants.MinMillisPerFrame * 1_000_000
	}

	s.renderTime += s.FrameNanos

	if s.NewFrame != nil {
		s.NewFrame()
	}

	for d := range s.Dynamics {
		d.NewFrame()
	}

	for s.renderTime >= constants.TimeStepNS {
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
		s.renderTime -= constants.TimeStepNS
		s.SimTimestamp += constants.TimeStepNS
	}

	// Update the blended values
	s.RenderStateBlend = float64(s.renderTime) / float64(constants.TimeStepNS)

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
		ID:           id,
		Timestamp:    s.Timestamp,
		SimTimestamp: s.SimTimestamp,
		Data:         data,
	})
}
