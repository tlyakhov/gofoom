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

func alcGetError(d Device) int32 {
	dev := (*C.ALCdevice)(d)
	return int32(C.alcGetError(dev))
}

func alcOpenDevice(name string) Device {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))

	return (Device)(C.alcOpenDevice((*C.ALCchar)(unsafe.Pointer(n))))
}

func alcCloseDevice(d Device) bool {
	dev := (*C.ALCdevice)(d)
	return C.alcCloseDevice(dev) == C.ALC_TRUE
}

func alcCreateContext(d Device, attrs []int32) unsafe.Pointer {
	dev := (*C.ALCdevice)(d)
	return (unsafe.Pointer)(C.alcCreateContext(dev, nil))
}

func alcMakeContextCurrent(c unsafe.Pointer) bool {
	ctx := (*C.ALCcontext)(c)
	return C.alcMakeContextCurrent(ctx) == C.ALC_TRUE
}

func alcDestroyContext(c unsafe.Pointer) {
	C.alcDestroyContext((*C.ALCcontext)(c))
}

func alcIsExtensionPresent(d Device, name string) bool {
	dev := (*C.ALCdevice)(d)
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	return C.alcIsExtensionPresent(dev, (*C.ALCchar)(unsafe.Pointer(n))) == C.ALC_TRUE
}
