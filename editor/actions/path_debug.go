// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

type PathDebug struct {
	Place

	Bla bool
}

func (a *PathDebug) Activate() {}

func (a *PathDebug) EndPoint() bool {
	if !a.Place.EndPoint() {
		return false
	}

	a.State().PathDebugStart = *a.WorldGrid(&a.State().MouseDownWorld)
	a.State().PathDebugEnd = *a.WorldGrid(&a.State().MouseWorld)

	a.ActionFinished(false, true, true)
	return true
}

func (a *PathDebug) Cancel() {
	a.ActionFinished(true, true, true)
}

func (a *PathDebug) Status() string {
	return ""
}
