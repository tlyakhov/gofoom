// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || (linux && !android) || windows

package al

/*
#cgo darwin   CFLAGS:  -DGOOS_darwin -I/opt/local/include
#cgo linux    CFLAGS:  -DGOOS_linux
#cgo windows  CFLAGS:  -DGOOS_windows
#cgo darwin   LDFLAGS: -lopenal -L/opt/local/lib
#cgo linux    LDFLAGS: -lopenal
#cgo windows  LDFLAGS: -lOpenAL32

#ifdef GOOS_darwin
#include <stdlib.h>
#include <AL/alc.h>
#endif

#ifdef GOOS_linux
#include <stdlib.h>
#include <AL/alc.h>
#endif

#ifdef GOOS_windows
#include <windows.h>
#include <stdlib.h>
#include <AL/alc.h>
#endif
*/
import "C"
import "unsafe"

/*
On Ubuntu 14.04 'Trusty', you may have to install these libraries:
sudo apt-get install libopenal-dev
*/

func alcGetError(d alDevice) int32 {
	dev := (*C.ALCdevice)(d)
	return int32(C.alcGetError(dev))
}

func alcOpenDevice(name string) alDevice {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))

	return (alDevice)(C.alcOpenDevice((*C.ALCchar)(unsafe.Pointer(n))))
}

func alcCloseDevice(d alDevice) bool {
	dev := (*C.ALCdevice)(d)
	return C.alcCloseDevice(dev) == C.ALC_TRUE
}

func alcCreateContext(d alDevice, attrs []int32) unsafe.Pointer {
	dev := (*C.ALCdevice)(d)

	var ptr *C.ALCint = nil
	if len(attrs) > 0 {
		zeroTerminated := append(make([]int32, 0, len(attrs)+1), attrs...)
		ptr = (*C.ALCint)(unsafe.Pointer(&zeroTerminated[0]))
	}
	return (unsafe.Pointer)(C.alcCreateContext(dev, ptr))
}

func alcMakeContextCurrent(c unsafe.Pointer) bool {
	ctx := (*C.ALCcontext)(c)
	return C.alcMakeContextCurrent(ctx) == C.ALC_TRUE
}

func alcDestroyContext(c unsafe.Pointer) {
	C.alcDestroyContext((*C.ALCcontext)(c))
}

func alcIsExtensionPresent(d alDevice, name string) bool {
	dev := (*C.ALCdevice)(d)
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	return C.alcIsExtensionPresent(dev, (*C.ALCchar)(unsafe.Pointer(n))) == C.ALC_TRUE
}

func alcGetIntegerv(d alDevice, k Enum, v []int32) {
	dev := (*C.ALCdevice)(d)
	C.alcGetIntegerv(dev, C.ALCenum(k), C.ALCsizei(len(v)), (*C.ALCint)(unsafe.Pointer(&v[0])))
}

func alcGetString(d alDevice, v Enum) string {
	dev := (*C.ALCdevice)(d)
	value := C.alcGetString(dev, C.ALCenum(v))
	return C.GoString((*C.char)(value))
}

func alcGetStrings(d alDevice, v Enum) []string {
	dev := (*C.ALCdevice)(d)
	// OpenAL has an incredibly dangerous way of returning a list of strings,
	// where each string is null-terminated, and the list ends with two nulls in
	// a row.
	value := C.alcGetString(dev, C.ALCenum(v))
	result := make([]string, 0)

	for {
		current := C.GoString((*C.char)(value))
		if len(current) == 0 {
			break
		}
		result = append(result, current)
		value = (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(value)) + uintptr(len(current)+1)))
	}
	return result
}
