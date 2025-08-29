package audio

import (
	"fmt"
	"os"
	"strings"
	"tlyakhov/gofoom/components/audio/al"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type Sound struct {
	ecs.Attached

	Source        string  `editable:"File" edit_type:"file"`
	ConvertToMono bool    `editable:"Convert To Mono"`
	Gain          float64 `editable:"Gain"`

	buffer al.Buffer
	bytes  []byte
	loaded bool
}

func (snd *Sound) String() string {
	return "Sound File"
}

func (snd *Sound) Load() error {
	if snd.loaded {
		al.DeleteBuffers(snd.buffer)
		snd.loaded = false
		snd.buffer = 0
		snd.bytes = nil
	}
	if snd.Source == "" {
		return nil
	}

	var mixer *Mixer
	// Ensure sound system is initialized
	if mixer = ecs.Singleton(MixerCID).(*Mixer); mixer == nil {
		return nil
	}

	f, err := os.Open(snd.Source)
	if err != nil {
		return err
	}
	defer f.Close()

	if strings.HasSuffix(snd.Source, ".ogg") {
		return snd.loadOgg(mixer, f)
	} else if strings.HasSuffix(snd.Source, ".wav") {
		return snd.loadWav(mixer, f)
	} else if strings.HasSuffix(snd.Source, ".mp3") {
		return snd.loadMP3(mixer, f)
	} else {
		return fmt.Errorf("Sound.Load: tried to load unknown (not wav, ogg) format audio file %v", snd.Source)
	}
}

func (snd *Sound) Construct(data map[string]any) {
	snd.Attached.Construct(data)
	snd.ConvertToMono = false
	snd.Gain = 1

	if data == nil {
		return
	}

	if v, ok := data["Source"]; ok {
		snd.Source = cast.ToString(v)
	}
	if v, ok := data["ConvertToMono"]; ok {
		snd.ConvertToMono = cast.ToBool(v)
	}
	if v, ok := data["Gain"]; ok {
		snd.Gain = cast.ToFloat64(v)
	}
}

func (snd *Sound) Serialize() map[string]any {
	result := snd.Attached.Serialize()

	result["Source"] = snd.Source
	if snd.ConvertToMono {
		result["ConvertToMono"] = snd.ConvertToMono
	}
	if snd.Gain != 1 {
		result["Gain"] = snd.Gain
	}

	return result
}
