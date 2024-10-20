// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import "tlyakhov/gofoom/concepts"

type LightmapCell struct {
	Light      concepts.Vector3
	Timestamp  uint64
	RandomSeed uint64
}
