// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || linux || windows

package al

import (
	"errors"
	"sync"
	"unsafe"
)

type Device unsafe.Pointer

var (
	mu      sync.Mutex
	device  Device
	context unsafe.Pointer
)

// DeviceError returns the last known error from the current device.
func DeviceError() int32 {
	return alcGetError(device)
}

// OpenDevice opens the default audio device.
// Calls to OpenDevice are safe for concurrent use.
func OpenDevice() (Device, error) {
	mu.Lock()
	defer mu.Unlock()

	// already opened
	if device != nil {
		return nil, nil
	}

	dev := alcOpenDevice("")
	if dev == nil {
		return nil, errors.New("al: cannot open the default audio device")
	}
	ctx := alcCreateContext(dev, nil)
	if ctx == nil {
		alcCloseDevice(dev)
		return nil, errors.New("al: cannot create a new context")
	}
	if !alcMakeContextCurrent(ctx) {
		alcCloseDevice(dev)
		return nil, errors.New("al: cannot make context current")
	}

	alLoadEAXProcs()

	device = dev
	context = ctx
	return dev, nil
}

// CloseDevice closes the device and frees related resources.
// Calls to CloseDevice are safe for concurrent use.
func CloseDevice() {
	mu.Lock()
	defer mu.Unlock()

	if device == nil {
		return
	}

	alcCloseDevice(device)
	if context != nil {
		alcDestroyContext(context)
	}
	device = nil
	context = nil
}

func IsExtensionPresent(d Device, name string) bool {
	return alcIsExtensionPresent(d, name)
}
