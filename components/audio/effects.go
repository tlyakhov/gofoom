package audio

// Effect represents a DSP effect that can be applied to audio.
type Effect interface {
	Process(data []int16)
}

// Delay effect.
type Delay struct {
	// Delay in samples
	Delay int
	// Feedback amount (0.0 to 1.0)
	Feedback float64
	// Wet/dry mix (0.0 to 1.0)
	Mix float64

	buffer []int16
	index  int
}

// NewDelay creates a new Reverb effect.
func NewDelay(delay int, feedback, mix float64) *Delay {
	return &Delay{
		Delay:    delay,
		Feedback: feedback,
		Mix:      mix,
		buffer:   make([]int16, delay),
	}
}

// Process applies the reverb effect to the audio data.
func (r *Delay) Process(data []int16) {
	for i, sample := range data {
		// Simple reverb logic
		delayedSample := r.buffer[r.index]
		r.buffer[r.index] = sample + int16(float64(delayedSample)*r.Feedback)
		data[i] = int16(float64(sample)*(1-r.Mix) + float64(delayedSample)*r.Mix)
		r.index = (r.index + 1) % len(r.buffer)
	}
}

// BitCrush effect.
type BitCrush struct {
	// Number of bits to crush to
	Bits int
}

// Process applies the bit crush effect.
func (bc *BitCrush) Process(data []int16) {
	step := 1 << (16 - bc.Bits)
	for i, sample := range data {
		data[i] = sample / int16(step) * int16(step)
	}
}
