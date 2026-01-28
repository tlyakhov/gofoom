// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package audio

import (
	"tlyakhov/gofoom/components/audio/al"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

// SoundEvent represents an active piece of audio. These are dynamically created when
// a sound is triggered by a source.
type SoundEvent struct {
	ecs.Attached

	SourceEntity ecs.Entity
	Sound        ecs.Entity
	Tag          string
	Offset       concepts.Vector3

	source al.Source
}

func (s *SoundEvent) String() string {
	return "Sound Event for " + s.SourceEntity.Format()
}

func (s *SoundEvent) Construct(data map[string]any) {
	s.Attached.Construct(data)
	s.Flags |= ecs.ComponentInternal // never serialize this
}

func (s *SoundEvent) Serialize() map[string]any {
	result := s.Attached.Serialize()

	return result
}

func (s *SoundEvent) SetPosition(v *concepts.Vector3) {
	p := al.Vector{float32((v[0] + s.Offset[0]) / constants.UnitsPerMeter),
		float32((v[1] + s.Offset[1]) / constants.UnitsPerMeter),
		float32((v[2] + s.Offset[2]) / constants.UnitsPerMeter)}
	s.source.SetPosition(p)
}

func (s *SoundEvent) SetVelocity(v *concepts.Vector3) {
	s.source.SetVelocity(alVectorFromUnits(v))
}

func (s *SoundEvent) SetGain(gain float32) {
	if snd := GetSound(s.Sound); snd != nil {
		gain *= float32(snd.Gain)
	}
	s.source.SetGain(gain)
}

func (s *SoundEvent) SetDirection(dir *concepts.Vector3) {
	s.source.SetDirection(al.Vector{float32(dir[0]), float32(dir[1]), float32(dir[2])})
}

func (s *SoundEvent) SetPitchMultiplier(p float64) {
	s.source.Setf(al.ParamPitch, float32(p))
}

// Stop halts playback of the sound.
func (s *SoundEvent) Stop() {
	al.StopSources(s.source)
}
