// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"math"
)

// TODO: Make separate 2D and 3D versions
type Ray struct {
	Start, End, Delta  Vector3
	Angle              float64
	AngleCos, AngleSin float64
	Pitch              float64
	Limit              float64
}

func (r *Ray) FromAngleAndLimit(a float64, pitch float64, limit float64) {
	if pitch > 90 {
		pitch = 90
	}
	if pitch < -90 {
		pitch = -90
	}

	r.Limit = limit
	r.Pitch = pitch

	r.Angle = a
	r.AngleSin, r.AngleCos = math.Sincos(a * Deg2rad)
	r.Delta[0] = r.AngleCos * math.Cos(pitch*Deg2rad)
	r.Delta[1] = r.AngleSin * math.Cos(pitch*Deg2rad)
	r.Delta[2] = math.Sin(pitch * Deg2rad)
	r.End[0] = r.Start[0] + limit*r.Delta[0]
	r.End[1] = r.Start[1] + limit*r.Delta[1]
	r.End[2] = r.Start[2] + limit*r.Delta[2]
}

func (r *Ray) AnglesFromStartEnd() {
	r.Delta[0] = r.End[0] - r.Start[0]
	r.Delta[1] = r.End[1] - r.Start[1]
	r.Delta[2] = r.End[2] - r.Start[2]
	r.Limit = r.Delta.Length()
	if r.Limit > 0 {
		r.Angle = math.Atan2(r.Delta[1], r.Delta[0]) * Rad2deg
		r.AngleSin, r.AngleCos = math.Sincos(r.Angle * Deg2rad)
		r.Delta.MulSelf(1.0 / r.Limit)
		r.Pitch = math.Asin(r.Delta[2]) * Rad2deg
	} else {
		r.Angle = 0
		r.AngleCos = 1
		r.AngleSin = 0
		r.Pitch = 0
	}
}

func (r *Ray) DistTo(p *Vector2) float64 {
	dx := math.Abs(p[0] - r.Start[0])
	dy := math.Abs(p[1] - r.Start[1])
	if dy > dx {
		return math.Abs(dy / r.AngleSin)
	}
	return math.Abs(dx / r.AngleCos)
}
