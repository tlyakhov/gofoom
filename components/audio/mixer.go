package audio

import (
	"fmt"
	"log"
	"sync"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"

	"tlyakhov/gofoom/components/audio/al"

	"github.com/kelindar/bitmap"
)

// Mixer manages the audio state and playback. It should be a singleton
type Mixer struct {
	ecs.Attached `ecs:"singleton"`

	// 44100, 48000, 96000 etc
	SampleRate int `editable:"Sample Rate"`
	// 2 for stereo, 6 for 5.1 surround
	Channels   int    `editable:"Channels"`
	DeviceName string `edtiable:"Device Name"`

	Error error // We should expose this somewhere

	formats     map[string]al.Enum
	events      map[al.Source]*SoundEvent
	usedSources bitmap.Bitmap
	sources     []al.Source
	fx          []al.Effect
	fxSlots     []al.AuxEffectSlot

	device *al.Device
	// Mutex for thread-safe operations
	Lock sync.Mutex
}

func (m *Mixer) String() string {
	return "Audio Mixer"
}

func (m *Mixer) deleteSoundEvent(source al.Source) {
	if event, ok := m.events[source]; ok {
		if event.IsAttached() {
			ecs.Delete(event.Entity)
		}
		delete(m.events, source)
	}
}

func (m *Mixer) paramsToFormat(channels int, bits int, isFloat bool) al.Enum {
	switch {
	case channels == 1 && bits == 8 && !isFloat:
		return al.FormatMono8
	case channels == 2 && bits == 8 && !isFloat:
		return al.FormatStereo8
	case channels == 1 && bits == 16 && !isFloat:
		return al.FormatMono16
	case channels == 2 && bits == 16 && !isFloat:
		return al.FormatStereo16
	case channels == 1 && bits == 32 && !isFloat:
		return m.formats["AL_FORMAT_MONO_I32"]
	case channels == 2 && bits == 32 && !isFloat:
		return m.formats["AL_FORMAT_STEREO_I32"]
	case channels == 1 && bits == 32 && isFloat:
		return m.formats["AL_FORMAT_MONO_FLOAT32"]
	case channels == 2 && bits == 32 && isFloat:
		return m.formats["AL_FORMAT_STEREO_FLOAT32"]
	}
	return 0
}

var extFormats = []string{
	"AL_FORMAT_QUAD16",
	"AL_FORMAT_MONO_FLOAT32",
	"AL_FORMAT_STEREO_FLOAT32",
	"AL_FORMAT_MONO_I32",
	"AL_FORMAT_STEREO_I32",
	"AL_FORMAT_REAR_I32",
	"AL_FORMAT_REAR_FLOAT32",
	"AL_FORMAT_QUAD_I32",
	"AL_FORMAT_QUAD_FLOAT32",
	"AL_FORMAT_51CHN_I32",
	"AL_FORMAT_51CHN_FLOAT32",
	"AL_FORMAT_61CHN_I32",
	"AL_FORMAT_61CHN_FLOAT32",
	"AL_FORMAT_71CHN_I32",
	"AL_FORMAT_71CHN_FLOAT32",
}

func (m *Mixer) Construct(data map[string]any) {
	m.Attached.Construct(data)

	m.Flags |= ecs.ComponentInternal // never serialize this
	m.Error = nil
	m.Channels = 2
	m.SampleRate = 48000
	m.DeviceName = ""
	m.formats = make(map[string]al.Enum)

	var err error
	if m.device, err = al.OpenDevice(m.DeviceName); err != nil {
		m.Error = fmt.Errorf("failed to open OpenAL device: %w", err)
		return
	}

	// Populate enum values for extended audio formats
	for _, f := range extFormats {
		m.formats[f] = al.GetEnumValue(f)
	}

	if !al.IsExtensionPresent("ALC_EXT_EFX") {
		m.Error = fmt.Errorf("EFX not supported")
		//CloseAL();
		return
	}

	numSends := m.device.GetIntegerv(al.MaxAuxiliarySends, 1)[0]
	if m.device.Error() != al.NoError || numSends < 2 {
		m.Error = fmt.Errorf("device does not support multiple sends (got %d, need 2)", numSends)
		//CloseAL()
		return
	}

	numVoices := 32
	m.sources = al.GenSources(numVoices)
	m.events = make(map[al.Source]*SoundEvent)
	m.usedSources.Grow(uint32(len(m.sources)))

	log.Printf("Initialized OpenAL audio: %vhz %v channels, %v voices, %v aux sends. Extensions: %v", m.SampleRate, m.Channels, len(m.sources), numSends, al.Extensions())
	log.Printf("Devices: %v", al.AllDevices())

	// Testing EAX reverb effects:
	// References:
	// https://github.com/kcat/openal-soft/blob/master/examples/almultireverb.c

	m.fx = al.GenEffects(1)
	m.fx[0].Load(al.EfxReverbPresets["Stonecorridor"])

	m.fxSlots = al.GenAuxEffectSlots(1)
	m.fxSlots[0].AuxiliaryEffectSloti(al.EffectSlotEffect, int32(m.fx[0]))
}

func (m *Mixer) SetReverbPreset(preset string) {
	if len(m.fx) == 0 {
		return
	}
	efx, ok := al.EfxReverbPresets[preset]
	if !ok {
		return
	}
	m.fx[0].Load(efx)
	m.fxSlots[0].AuxiliaryEffectSloti(al.EffectSlotEffect, int32(m.fx[0]))
}

func (m *Mixer) Serialize() map[string]any {
	result := m.Attached.Serialize()

	return result
}

// Close shuts down the audio engine.
func (m *Mixer) OnDelete() {
	for _, event := range m.events {
		if event == nil {
			continue
		}
		event.Stop()
	}
	m.events = nil
	al.DeleteSources(m.sources...)
	al.DeleteEffects(m.fx...)
	m.device.Close()
}

func (m *Mixer) play(snd *Sound) (al.Source, error) {
	if snd == nil {
		return 0, nil
	}
	m.Lock.Lock()
	defer m.Lock.Unlock()

	voice, ok := m.usedSources.MinZero()
	if !ok || int(voice) >= len(m.sources) {
		return 0, fmt.Errorf("audio.Mixer: Ran out of voices, can't play %v", snd)
	}

	source := m.sources[voice]
	// log.Printf("Playing %v on source %v", snd.Source, source)
	source.Set3i(al.AuxiliarySendFilter, int32(m.fxSlots[0]), 0, al.FilterNull)
	source.SetGain(float32(snd.Gain))
	source.SetBuffer(snd.buffer)
	source.Setf(al.ParamPitch, 1)
	// source.QueueBuffers(snd.buffer)
	//if source.Geti(al.ParamSourceState) != al.Playing {
	al.PlaySources(source)
	//}
	m.usedSources.Set(voice)

	return source, nil
}

func (m *Mixer) PollSources() {
	m.usedSources.Range(func(index uint32) {
		src := m.sources[index]
		state := src.Geti(al.ParamSourceState)
		if state == al.Stopped {
			m.deleteSoundEvent(src)
			m.usedSources.Remove(index)
		}
	})

}

func (m *Mixer) SetListenerPosition(v *concepts.Vector3) {
	al.SetListenerPosition(alVector(v))
}

var alUpVector = alVector(&concepts.Vector3{0, 0, constants.UnitsPerMeter})

func (m *Mixer) SetListenerOrientation(v *concepts.Vector3) {
	al.SetListenerOrientation(al.Orientation{Forward: alVector(v), Up: alUpVector})
}

func (m *Mixer) SetListenerVelocity(v *concepts.Vector3) {
	al.SetListenerVelocity(alVector(v))
}

// TODO: Add hysteresis rather than just the onePerTag param
// PlaySound initiates playback for a sound asset, optionally attached to a
// source (e.g. a body, a sector)
func PlaySound(sound ecs.Entity, sourceEntity ecs.Entity, tag string, onePerTag bool) (*SoundEvent, error) {
	m := ecs.Singleton(MixerCID).(*Mixer)
	if m == nil {
		return nil, nil
	}
	snd := GetSound(sound)
	if snd == nil {
		return nil, nil
	}

	if onePerTag {
		arena := ecs.ArenaFor[SoundEvent](SoundEventCID)
		for i := range arena.Cap() {
			event := arena.Value(i)
			if event == nil {
				continue
			}
			if event.Tag == tag {
				//log.Printf("Already playing %v with tag %v", snd.Source, tag)
				return event, nil
			}
		}
	}

	var source al.Source
	var err error
	if source, err = m.play(snd); err != nil {
		return nil, err
	}
	event := ecs.NewAttachedComponent(ecs.NewEntity(), SoundEventCID).(*SoundEvent)
	event.Sound = snd.Entity
	event.SourceEntity = sourceEntity
	event.source = source
	event.Tag = tag
	m.deleteSoundEvent(source)
	m.events[source] = event
	return event, nil
}

func alVector(v *concepts.Vector3) al.Vector {
	return al.Vector{float32(v[0] / constants.UnitsPerMeter),
		float32(v[1] / constants.UnitsPerMeter),
		float32(v[2] / constants.UnitsPerMeter)}
}
func SetReverbPreset(preset string) {
	m := ecs.Singleton(MixerCID).(*Mixer)
	if m == nil {
		return
	}
	m.SetReverbPreset(preset)
}
