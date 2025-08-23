package audio

import (
	"fmt"
	"strings"
	"tlyakhov/gofoom/components/audio/al"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type Sound struct {
	ecs.Attached

	Source        string `editable:"File" edit_type:"file"`
	ConvertToMono bool   `editable:"Convert To Mono"`

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
	if !strings.HasSuffix(snd.Source, ".wav") {
		return fmt.Errorf("Sound.Load: tried to load non-wav format audio file %v", snd.Source)
	}

	return snd.loadWav()
}

func (snd *Sound) Construct(data map[string]any) {
	snd.Attached.Construct(data)
	snd.ConvertToMono = false

	if data == nil {
		return
	}

	if v, ok := data["Source"]; ok {
		snd.Source = cast.ToString(v)
	}
	if v, ok := data["ConvertToMono"]; ok {
		snd.ConvertToMono = cast.ToBool(v)
	}
}

func (snd *Sound) Serialize() map[string]any {
	result := snd.Attached.Serialize()

	result["Source"] = snd.Source
	if snd.ConvertToMono {
		result["ConvertToMono"] = snd.ConvertToMono
	}

	return result
}
