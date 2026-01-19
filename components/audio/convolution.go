// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package audio

import (
	"fmt"
	"math"
	"tlyakhov/gofoom/ecs"

	"gonum.org/v1/gonum/dsp/fourier"
)

/* TODO: This is all too slow and doesn't work. Probably delete all of this
 */

type convReverbChannel struct {
	impulseResponseFFT []complex128 // Pre-computed FFT of the impulse response
	fft                *fourier.FFT // The FFT transformer instance
	overlap            []float64    // Buffer for the overlap-add method
}

// ConvReverb applies a reverb effect using FFT-based convolution.
// It implements the Effect interface using the Overlap-Add method.
type ConvReverb struct {
	fftSize     int // The size of the FFT operations
	blockSize   int // The size of the input audio blocks to process
	channels    []convReverbChannel
	inputFloat  []float64
	outputFloat []float64
}

// nextPowerOfTwo calculates the next power of two for a given integer.
// This is important for optimizing FFT performance.
func nextPowerOfTwo(n int) int {
	return int(math.Pow(2, math.Ceil(math.Log2(float64(n)))))
}

// NewConvReverb creates a new FFT-based convolution reverb effect.
// impulseResponsePath: Path to the mono 16-bit WAV impulse response file.
// blockSize: The size of the audio blocks to process at a time.
func NewConvReverb(irEntity ecs.Entity, blockSize int) (*ConvReverb, error) {
	// --- 1. Load Impulse Response from WAV file ---
	snd := GetSound(irEntity)
	if snd == nil {
		return nil, fmt.Errorf("no IR")
	}

	result := &ConvReverb{
		channels:  make([]convReverbChannel, Mixer.Channels),
		blockSize: blockSize,
	}

	sndSamples := snd.bytes //snd.GetAudioData()
	irSamples := make([]float64, len(sndSamples))
	for i := range irSamples {
		irSamples[i] = float64(sndSamples[i]) / 32768.0
	}
	// --- 2. Setup FFT parameters ---
	irSize := len(irSamples) / Mixer.Channels
	// The FFT size must be large enough to hold the convolution result without aliasing.
	// Convolution length = blockSize + irSize - 1
	result.fftSize = nextPowerOfTwo(blockSize + irSize - 1)
	for i := range result.channels {
		result.channels[i].overlap = make([]float64, result.fftSize-blockSize)
		result.channels[i].fft = fourier.NewFFT(result.fftSize)
		// --- 3. Pre-compute the FFT of the Impulse Response ---
		// Zero-pad the impulse response to the FFT size.
		paddedIR := make([]float64, result.fftSize)
		copy(paddedIR, irSamples[i*irSize:])
		// Compute the FFT and store it. This is the "H(w)" in the frequency domain.
		result.channels[i].impulseResponseFFT = result.channels[i].fft.Coefficients(nil, paddedIR)
	}
	return result, nil
}

// Process applies the convolution reverb to the input audio data.
func (c *ConvReverb) Process(data []int16, channels int) {
	if len(c.inputFloat) != len(data) {
		c.inputFloat = make([]float64, len(data))
		c.outputFloat = make([]float64, len(data))
	}

	// Convert input int16 buffer to a float64 buffer for processing.
	for i, s := range data {
		c.inputFloat[i] = float64(s) / 32768.0
	}

	chSize := len(c.inputFloat) / len(c.channels)
	for ch := range c.channels {
		// Process the input signal in blocks of size `c.blockSize`.
		for i := 0; i < chSize; i += c.blockSize {
			// Determine the end of the current block.
			end := min(i+c.blockSize, chSize)
			block := c.inputFloat[i+chSize*ch : end+chSize*ch]

			// --- Overlap-Add Algorithm Step by Step ---

			// 1. Zero-pad the current block to the FFT size.
			paddedBlock := make([]float64, c.fftSize)
			copy(paddedBlock, block)

			// 2. Compute the FFT of the padded input block -> X(w).
			_ = c.channels[ch].fft.Coefficients(nil, paddedBlock)

			/*	// 3. Perform element-wise multiplication in the frequency domain: Y(w) = H(w) * X(w).
				convolvedFFT := make([]complex128, len(blockFFT))
				for j := range blockFFT {
					convolvedFFT[j] = c.channels[ch].impulseResponseFFT[j] * blockFFT[j]
				}

				// 4. Compute the Inverse FFT to get the result back in the time domain.
				convolvedBlock := c.channels[ch].fft.Sequence(nil, convolvedFFT)

				// 4a. *** FIX: Scale the result of the IFFT. ***
				// The IFFT is unnormalized, so we must divide by the FFT size
				// to get the correct amplitude for the convolution result.
				scale := float64(c.fftSize)
				for j := range convolvedBlock {
					convolvedBlock[j] /= scale
				}

				// 5. Add the overlap from the *previous* block's processing.
				for j := 0; j < len(c.channels[ch].overlap) && j < len(convolvedBlock); j++ {
					convolvedBlock[j] += c.channels[ch].overlap[j]
				}*/

			// 6. Copy the first `blockSize` samples of the result to our final output buffer.
			//outputSize := end - i
			//copy(outputFloat[i+chSize*ch:end+chSize*ch], convolvedBlock[:outputSize])

			copy(c.outputFloat[i+chSize*ch:end+chSize*ch], c.inputFloat[i+chSize*ch:end+chSize*ch])

			// 7. Save the "tail" of the convolution result as the new overlap for the *next* block.
			//copy(c.channels[ch].overlap, convolvedBlock[c.blockSize:])
		}
	}

	// Convert the processed float64 buffer back to int16, clamping values.
	for i := range data {
		sample := c.outputFloat[i] * 32768.0
		if sample > math.MaxInt16 {
			data[i] = math.MaxInt16
		} else if sample < math.MinInt16 {
			data[i] = math.MinInt16
		} else {
			data[i] = int16(sample)
		}
	}
}
