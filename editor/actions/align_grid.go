// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/concepts"
)

type AlignGrid struct {
	Place
	PrevA, PrevB, A, B concepts.Vector2
}

func (a *AlignGrid) Activate() {}

func (a *AlignGrid) EndPoint() bool {
	if !a.Place.EndPoint() {
		return false
	}
	a.State().Lock.Lock()

	a.PrevA, a.PrevB = a.State().GridA, a.State().GridB
	a.A = *a.WorldGrid(&a.State().MouseDownWorld)
	a.B = *a.WorldGrid(&a.State().MouseWorld)
	if a.A.DistSq(&a.B) < 0.001 {
		a.A = concepts.Vector2{}
		a.B = concepts.Vector2{0, 1}
	}
	a.State().GridA, a.State().GridB = a.A, a.B
	a.State().Lock.Unlock()
	a.ActionFinished(false, false, false)
	return true
}

func (a *AlignGrid) Cancel() {
	a.ActionFinished(true, false, false)
}

func (a *AlignGrid) Status() string {
	return "Click to align grid"
}

func (a *AlignGrid) AffectsWorld() bool { return false }
