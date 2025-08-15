package audio

import (
	"sync"

	"github.com/veandco/go-sdl2/mix"
)

// SoundEvent represents an active piece of audio. These are dynamically created when
// a sound is triggered by a source.
type SoundEvent struct {
	chunk *mix.Chunk
	// Channel the sound is playing on
	channel int
	// Effects applied to this sound
	Effects []Effect
	// Mutex for thread-safe operations
	mu sync.Mutex
}

// SetVolume adjusts the volume of the sound.
func (s *SoundEvent) SetVolume(volume float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.chunk != nil {
		mix.Volume(s.channel, int(volume*float64(mix.MAX_VOLUME)))
	}
}

// SetVolume adjusts the volume of the sound.
func (s *SoundEvent) SetPosition(angle, distance float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.chunk != nil {
		mix.SetPosition(s.channel, int16(angle), uint8(distance*0.1))
	}
}

// Stop halts playback of the sound.
func (s *SoundEvent) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.chunk != nil {
		mix.HaltChannel(s.channel)
	}
}
