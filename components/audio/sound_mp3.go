package audio

import (
	"io"
	"os"
	"tlyakhov/gofoom/components/audio/al"

	gomp3 "github.com/hajimehoshi/go-mp3"
)

func (snd *Sound) loadMP3(mixer *Mixer, f *os.File) error {
	d, err := gomp3.NewDecoder(f)
	if err != nil {
		return err
	}

	incomingChannels := 2 // Always?
	outgoingChannels := incomingChannels
	if snd.ConvertToMono {
		outgoingChannels = 1
	}

	// TODO: Provide way to stream files to memory
	snd.bytes = make([]byte, 0)
	raw := make([]byte, 8192)
	for {
		n, err := d.Read(raw)

		i := 0
		for i < n {
			// go-mp3 spits out 16bit stereo samples
			snd.bytes = append(snd.bytes, raw[i], raw[i+1])
			i += 2
			if snd.ConvertToMono && incomingChannels > 1 {
				i += (int(incomingChannels) - 1) * 2
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	snd.buffer = al.GenBuffers(1)[0]
	format := mixer.paramsToFormat(outgoingChannels, 16, false)
	snd.buffer.BufferData(format, snd.bytes, int32(d.SampleRate()))
	snd.loaded = true
	return nil
}
