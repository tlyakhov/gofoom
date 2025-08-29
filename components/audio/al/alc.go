// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || linux || windows

package al

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

type alDevice unsafe.Pointer
type Device struct {
	Name string

	ctx    unsafe.Pointer
	device alDevice
}

var globalLock sync.Mutex

// DeviceError returns the last known error from the current device.
func (d *Device) Error() int32 {
	return alcGetError(d.device)
}

// OpenDevice opens an audio device.
func OpenDevice(name string) (*Device, error) {
	globalLock.Lock()
	defer globalLock.Unlock()

	result := &Device{
		Name: name,
	}
	result.device = alcOpenDevice(name)
	if result.device == nil {
		return nil, fmt.Errorf("al: cannot open the audio device %v", name)
	}
	result.ctx = alcCreateContext(result.device, nil)
	if result.ctx == nil {
		alcCloseDevice(result.device)
		return nil, errors.New("al: cannot create a new context")
	}
	if !alcMakeContextCurrent(result.ctx) {
		alcCloseDevice(result.device)
		return nil, errors.New("al: cannot make context current")
	}

	alLoadEAXProcs()
	return result, nil
}

// Close closes the device and frees related resources.
// Calls are safe for concurrent use.
func (d *Device) Close() {
	globalLock.Lock()
	defer globalLock.Unlock()

	if d.device == nil {
		return
	}

	alcCloseDevice(d.device)
	d.device = nil

	if d.ctx != nil {
		alcDestroyContext(d.ctx)
		d.ctx = nil
	}
}

func IsExtensionPresent(name string) bool {
	return alcIsExtensionPresent(nil, name)
}

func AllDevices() []string {
	return alcGetStrings(nil, AlcAllDevicesSpecifier)
}

func (d *Device) GetIntegerv(k Enum, num int) []int32 {
	v := make([]int32, num)
	alcGetIntegerv(d.device, k, v)
	return v
}
