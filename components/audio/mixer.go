package audio

import (
	"fmt"
	"sync"
	"tlyakhov/gofoom/ecs"
	"unsafe"

	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

// Mixer manages the audio state and playback. It should be a singleton
type Mixer struct {
	ecs.Attached `ecs:"singleton"`

	// 44100, 48000, 96000 etc
	SampleRate int `editable:"Sample Rate"`
	// 2 for stereo, 6 for 5.1 surround
	Channels int `editable:"Channels"`

	// Effects can be applied to the master output
	MasterEffects []Effect
	Error         error // We should expose this somewhere

	spec     *sdl.AudioSpec
	channels []*SoundEvent
	// Mutex for thread-safe operations
	mu sync.Mutex
}

func (m *Mixer) String() string {
	return "Audio Mixer"
}

func (m *Mixer) Construct(data map[string]any) {
	m.Attached.Construct(data)

	m.Flags |= ecs.ComponentInternal // never serialize this
	m.Error = nil
	m.Channels = 2
	m.SampleRate = 48000
	m.spec = nil

	if err := sdl.Init(sdl.INIT_AUDIO); err != nil {
		m.Error = fmt.Errorf("could not initialize SDL: %v", err)
		return
	}

	if err := mix.Init(mix.INIT_MP3 | mix.INIT_FLAC); err != nil {
		m.Error = fmt.Errorf("could not initialize mixer: %v", err)
		return
	}

	m.spec = &sdl.AudioSpec{
		Freq:     int32(m.SampleRate),
		Format:   sdl.AUDIO_S16SYS,
		Channels: uint8(m.Channels),
		Samples:  4096,
	}

	if err := sdl.OpenAudio(m.spec, m.spec); err != nil {
		m.Error = fmt.Errorf("could not open audio: %v", err)
		return
	}

	if err := mix.OpenAudio(int(m.spec.Freq), uint16(m.spec.Format),
		int(m.spec.Channels), int(m.spec.Samples)); err != nil {
		m.Error = fmt.Errorf("could not open audio: %v", err)
		return
	}

	numChannels := mix.AllocateChannels(16)
	m.channels = make([]*SoundEvent, numChannels)
}

func (m *Mixer) Serialize() map[string]any {
	result := m.Attached.Serialize()

	return result
}

// SetChannels sets the number of output channels (e.g., 2 for stereo, 6 for 5.1).
func (m *Mixer) SetChannels(channels uint8) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mix.CloseAudio()
	m.spec.Channels = channels
	if err := mix.OpenAudio(int(m.spec.Freq), sdl.AUDIO_S16, int(m.spec.Channels), int(m.spec.Samples)); err != nil {
		return fmt.Errorf("could not reopen audio with %d channels: %v", channels, err)
	}
	return nil
}

// Close shuts down the audio engine.
func (m *Mixer) Close() {
	for _, event := range m.channels {
		event.Stop()
	}
	m.channels = nil
	mix.CloseAudio()
	mix.Quit()
	sdl.Quit()
}

func (m *Mixer) fxCallback(channel int, stream []byte) {
	e := m.channels[channel]
	if e == nil {
		return
	}
	converted := unsafe.Slice((*int16)(unsafe.Pointer(&stream[0])), len(stream)/2)

	for _, fx := range e.Effects {
		fx.Process(converted)
	}
}

func (m *Mixer) Play(snd *Sound, fx []Effect, volume float64) (*SoundEvent, error) {
	if snd == nil {
		return nil, nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if snd.Stream {
		if err := snd.Music.Play(-1); err != nil {
			return nil, err
		}
		return nil, nil
	} else {
		channel, err := snd.Chunk.Play(-1, 0)
		if err != nil {
			return nil, err
		}
		event := &SoundEvent{chunk: snd.Chunk, channel: channel, Effects: fx}
		m.channels[event.channel] = event
		mix.UnregisterAllEffects(channel)
		mix.RegisterEffect(channel, m.fxCallback, func(channel int) {})
		event.SetVolume(volume)

		return event, nil
	}
}

// Convenience function for scripts
func PlaySound(e ecs.Entity) (*SoundEvent, error) {
	m := ecs.Singleton(MixerCID).(*Mixer)
	if m == nil || m.spec == nil {
		return nil, nil
	}
	snd := GetSound(e)
	if snd == nil {
		return nil, nil
	}
	return m.Play(snd, nil, 1.0)
}
