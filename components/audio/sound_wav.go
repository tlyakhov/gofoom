package audio

import (
	"fmt"
	"math"
	"os"
	"tlyakhov/gofoom/components/audio/al"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

const (
	// From https://www.mmsp.ece.mcgill.ca/Documents/AudioFormats/WAVE/WAVE.html
	WavFormatPCM       = 0x0001
	WavFormatIEEEFloat = 0x0003
	WavFormatALaw      = 0x0006
	WavFormatMuLaw     = 0x0007
)

func intSampleToBytes(sample int, incomingBytes uint16, asFloat bool, bytes []byte) {
	if incomingBytes <= 3 && asFloat {
		panic("audio.intSampleToBytes: tried to convert float sample <32bits")
	}
	switch incomingBytes {
	case 1:
		bytes[0] = byte(sample & 0xFF)
	case 2:
		bytes[0] = byte(sample & 0xFF)
		bytes[1] = byte((sample >> 8) & 0xFF)
	case 3:
		// OpenAL doesn't support 24bit buffer format, let's promote to
		// 32bit float
		f := float32(float64(sample) / float64(1<<23-1))
		bits := math.Float32bits(f)
		bytes[0] = byte(bits & 0xFF)
		bytes[1] = byte((bits >> 8) & 0xFF)
		bytes[2] = byte((bits >> 16) & 0xFF)
		bytes[3] = byte((bits >> 24) & 0xFF)
	case 4:
		var bits uint32
		if asFloat {
			bits = uint32(sample)
		} else {
			// OpenAL 32bit int support isn't always there, let's promote to
			// 32bit float
			f := float32(float64(sample) / float64(math.MaxInt32))
			bits = math.Float32bits(f)
		}
		bytes[0] = byte(bits & 0xFF)
		bytes[1] = byte((bits >> 8) & 0xFF)
		bytes[2] = byte((bits >> 16) & 0xFF)
		bytes[3] = byte((bits >> 24) & 0xFF)
	}
}

func (snd *Sound) loadWav(mixer *Mixer, f *os.File) error {
	d := wav.NewDecoder(f)
	d.ReadMetadata()
	d.Rewind()
	// TODO: Find a better wav library or write my own that works better
	// with OpenAL
	incomingBytes := (d.BitDepth-1)/8 + 1
	outgoingBytes := incomingBytes
	if outgoingBytes == 3 {
		// OpenAL 32bit int support isn't always there, let's promote to
		// 32bit float
		outgoingBytes = 4
	}
	outgoingChannels := int(d.NumChans)
	if snd.ConvertToMono {
		outgoingChannels = 1
	}
	// TODO: Provide way to stream files to memory
	snd.bytes = make([]byte, 0)
	buf := &audio.IntBuffer{Data: make([]int, 4096)}
	index := 0
	for {
		n, err := d.PCMBuffer(buf)
		if err != nil {
			return fmt.Errorf("error loading file: %v", err)
		}
		if n == 0 {
			break
		}
		snd.bytes = append(snd.bytes, make([]byte, n*int(outgoingBytes)*outgoingChannels/int(d.NumChans))...)
		i := 0
		for i < n {
			intSampleToBytes(buf.Data[i], incomingBytes, d.WavAudioFormat == WavFormatIEEEFloat, snd.bytes[index:])
			index += int(outgoingBytes)
			i++
			if snd.ConvertToMono && d.NumChans > 1 {
				i += int(d.NumChans) - 1
			}
		}
	}

	snd.buffer = al.GenBuffers(1)[0]
	format := mixer.paramsToFormat(outgoingChannels, int(outgoingBytes)*8, outgoingBytes == 4)
	snd.buffer.BufferData(format, snd.bytes, int32(d.SampleRate))
	snd.loaded = true

	return nil
}
