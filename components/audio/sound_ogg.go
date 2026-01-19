// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package audio

import (
	"io"
	"log"
	"math"
	"os"
	"tlyakhov/gofoom/components/audio/al"

	"github.com/jfreymuth/oggvorbis"
)

func (snd *Sound) loadOgg(f *os.File) error {
	r, err := oggvorbis.NewReader(f)
	if err != nil {
		return err
	}

	log.Println(r.SampleRate())
	log.Println(r.Channels())

	incomingChannels := r.Channels()
	outgoingChannels := incomingChannels
	if snd.ConvertToMono {
		outgoingChannels = 1
	}

	// TODO: Provide way to stream files to memory
	snd.bytes = make([]byte, 0)
	floats := make([]float32, 8192)
	for {
		n, err := r.Read(floats)

		i := 0
		for i < n {
			sample := floats[i]
			bits := math.Float32bits(sample)
			snd.bytes = append(snd.bytes,
				byte(bits&0xFF), byte((bits>>8)&0xFF),
				byte((bits>>16)&0xFF), byte((bits>>24)&0xFF))
			i++
			if snd.ConvertToMono && incomingChannels > 1 {
				i += int(incomingChannels) - 1
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
	Mixer.buffers = append(Mixer.buffers, snd.buffer)
	format := Mixer.paramsToFormat(outgoingChannels, 32, true)
	snd.buffer.BufferData(format, snd.bytes, int32(r.SampleRate()))
	snd.loaded = true

	return nil
}
