// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"math"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type Ray struct {
	Start, End, Delta concepts.Vector2
	Angle             float64
	AngleCos          float64
	AngleSin          float64
}

func (r *Ray) Set(a float64) {
	r.Angle = a
	r.AngleSin, r.AngleCos = math.Sincos(a)
	r.End = concepts.Vector2{
		r.Start[0] + constants.MaxViewDistance*r.AngleCos,
		r.Start[1] + constants.MaxViewDistance*r.AngleSin,
	}
	r.Delta.From(&r.End).SubSelf(&r.Start)
}

func (r *Ray) AnglesFromStartEnd() {
	r.Delta.From(&r.End).SubSelf(&r.Start)
	r.Angle = math.Atan2(r.Delta[1], r.Delta[0])
	r.AngleSin, r.AngleCos = math.Sincos(r.Angle)
}

func (r *Ray) DistTo(p *concepts.Vector2) float64 {
	dx := math.Abs(p[0] - r.Start[0])
	dy := math.Abs(p[1] - r.Start[1])
	if dy > dx {
		return math.Abs(dy / r.AngleSin)
	}
	return math.Abs(dx / r.AngleCos)
}
