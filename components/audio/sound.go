package audio

import (
	"fmt"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
	"github.com/veandco/go-sdl2/mix"
)

type Sound struct {
	ecs.Attached

	Source string `editable:"File" edit_type:"file"`
	Stream bool   `editable:"Stream"`

	Chunk *mix.Chunk
	Music *mix.Music
}

func (snd *Sound) String() string {
	return "Sound File"
}

func (snd *Sound) Load() error {
	if snd.Chunk != nil {
		snd.Chunk.Free()
		snd.Chunk = nil
	}
	if snd.Music != nil {
		snd.Music.Free()
		snd.Music = nil
	}
	if snd.Source == "" {
		return nil
	}

	var err error
	if snd.Stream {
		snd.Music, err = mix.LoadMUS(snd.Source)
		if err != nil {
			return fmt.Errorf("could not load music: %v", err)
		}
	} else {
		snd.Chunk, err = mix.LoadWAV(snd.Source)
		if err != nil {
			return fmt.Errorf("could not load sound effect: %v", err)
		}
	}
	return nil
}

func (snd *Sound) Construct(data map[string]any) {
	snd.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Source"]; ok {
		snd.Source = cast.ToString(v)
	}

	if v, ok := data["Stream"]; ok {
		snd.Stream = cast.ToBool(v)
	}
}

func (snd *Sound) Serialize() map[string]any {
	result := snd.Attached.Serialize()

	result["Source"] = snd.Source
	result["Stream"] = snd.Stream

	return result
}
