// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package al

// Source represents an individual sound source in 3D-space.
// They take PCM data, apply modifications and then submit them to
// be mixed according to their spatial location.
type Source uint32

// GenSources generates n new sources. These sources should be deleted
// once they are not in use.
func GenSources(n int) []Source {
	return alGenSources(n)
}

// PlaySources plays the sources.
func PlaySources(source ...Source) {
	alSourcePlayv(source)
}

// PauseSources pauses the sources.
func PauseSources(source ...Source) {
	alSourcePausev(source)
}

// StopSources stops the sources.
func StopSources(source ...Source) {
	alSourceStopv(source)
}

// RewindSources rewinds the sources to their beginning positions.
func RewindSources(source ...Source) {
	alSourceRewindv(source)
}

// DeleteSources deletes the sources.
func DeleteSources(source ...Source) {
	if len(source) == 0 {
		return
	}
	alDeleteSources(source)
}

// Gain returns the source gain.
func (s Source) Gain() float32 {
	return s.Getf(paramGain)
}

// SetGain sets the source gain.
func (s Source) SetGain(v float32) {
	s.Setf(paramGain, v)
}

// MinGain returns the source's minimum gain setting.
func (s Source) MinGain() float32 {
	return s.Getf(paramMinGain)
}

// SetMinGain sets the source's minimum gain setting.
func (s Source) SetMinGain(v float32) {
	s.Setf(paramMinGain, v)
}

// MaxGain returns the source's maximum gain setting.
func (s Source) MaxGain() float32 {
	return s.Getf(paramMaxGain)
}

// SetMaxGain sets the source's maximum gain setting.
func (s Source) SetMaxGain(v float32) {
	s.Setf(paramMaxGain, v)
}

// Position returns the position of the source.
func (s Source) Position() Vector {
	v := Vector{}
	s.Getfv(paramPosition, v[:])
	return v
}

// SetPosition sets the position of the source.
func (s Source) SetPosition(v Vector) {
	s.Setfv(paramPosition, v[:])
}

// Velocity returns the source's velocity.
func (s Source) Velocity() Vector {
	v := Vector{}
	s.Getfv(paramVelocity, v[:])
	return v
}

// SetVelocity sets the source's velocity.
func (s Source) SetVelocity(v Vector) {
	s.Setfv(paramVelocity, v[:])
}

// Orientation returns the orientation of the source.
func (s Source) Orientation() Orientation {
	v := make([]float32, 6)
	s.Getfv(paramOrientation, v)
	return orientationFromSlice(v)
}

// SetOrientation sets the orientation of the source.
func (s Source) SetOrientation(o Orientation) {
	s.Setfv(paramOrientation, o.slice())
}

// State returns the playing state of the source.
func (s Source) State() int32 {
	return s.Geti(ParamSourceState)
}

// BuffersQueued returns the number of the queued buffers.
func (s Source) BuffersQueued() int32 {
	return s.Geti(paramBuffersQueued)
}

// BuffersProcessed returns the number of the processed buffers.
func (s Source) BuffersProcessed() int32 {
	return s.Geti(paramBuffersProcessed)
}

// OffsetSeconds returns the current playback position of the source in seconds.
func (s Source) OffsetSeconds() int32 {
	return s.Geti(paramSecOffset)
}

// OffsetSample returns the sample offset of the current playback position.
func (s Source) OffsetSample() int32 {
	return s.Geti(paramSampleOffset)
}

// OffsetByte returns the byte offset of the current playback position.
func (s Source) OffsetByte() int32 {
	return s.Geti(paramByteOffset)
}

// Geti returns the int32 value of the given parameter.
func (s Source) Geti(param int) int32 {
	return alGetSourcei(s, param)
}

// Getf returns the float32 value of the given parameter.
func (s Source) Getf(param int) float32 {
	return alGetSourcef(s, param)
}

// Getfv returns the float32 vector value of the given parameter.
func (s Source) Getfv(param int, v []float32) {
	alGetSourcefv(s, param, v)
}

// Seti sets an int32 value for the given parameter.
func (s Source) Seti(param int, v int32) {
	alSourcei(s, param, v)
}

// Seti sets 3 int32 values for the given parameter.
func (s Source) Set3i(param int, v1, v2, v3 int32) {
	alSource3i(s, param, v1, v2, v3)
}

// Setf sets a float32 value for the given parameter.
func (s Source) Setf(param int, v float32) {
	alSourcef(s, param, v)
}

// Setfv sets a float32 vector value for the given parameter.
func (s Source) Setfv(param int, v []float32) {
	alSourcefv(s, param, v)
}

// QueueBuffers adds the buffers to the buffer queue.
func (s Source) QueueBuffers(buffer ...Buffer) {
	alSourceQueueBuffers(s, buffer)
}

// UnqueueBuffers removes the specified buffers from the buffer queue.
func (s Source) UnqueueBuffers(buffer ...Buffer) {
	alSourceUnqueueBuffers(s, buffer)
}

func (s Source) SetBuffer(buffer Buffer) {
	alSourcei(s, paramBuffer, int32(buffer))
}
