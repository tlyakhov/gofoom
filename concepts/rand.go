// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"math"
	"time"
)

func RngXorShift64(xorSeed uint64) uint64 {
	if xorSeed == 0 {
		xorSeed = uint64(time.Now().UnixNano())
	}
	x := xorSeed * uint64(2685821657736338717)
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	/*x ^= x << 13
	x ^= x >> 7
	x ^= x << 17*/
	return x
}

func RngDecide(r uint64, mod uint64) bool {
	block := math.MaxUint64 / mod
	return r < block
}

// From https://gist.github.com/badboy/6267743
func Hash64to32(key uint64) uint32 {
	key = (^key) + (key << 18) // key = (key << 18) - key - 1;
	key = key ^ (key >> 31)
	key = key * 21 // key = (key + (key << 2)) + (key << 4);
	key = key ^ (key >> 11)
	key = key + (key << 6)
	key = key ^ (key >> 22)
	return uint32(key)
}
