// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

func NpcMove(body *core.Body, speed float64, delta *concepts.Vector3, face bool) {
	if face {
		angle := math.Atan2(delta[1], delta[0]) * concepts.Rad2deg
		if body.Angle.Procedural {
			body.Angle.Input = angle
		} else {
			body.Angle.Now = angle
		}
	}

	force := &concepts.Vector3{delta[0] * speed, delta[1] * speed, delta[2] * speed}
	mobile := core.GetMobile(body.Entity)
	switch {
	case mobile != nil:
		// TODO: Is this a hack?
		if !body.OnGround {
			force.MulSelf(0.01)
		}
		mobile.Force.AddSelf(force)
	case body.Pos.Procedural:
		body.Pos.Input.AddSelf(force.MulSelf(constants.TimeStepS))
		var bc BodyController
		if bc.Target(body, body.Entity) {
			bc.findBodySector()
		}
	default:
		body.Pos.Now.AddSelf(force.MulSelf(constants.TimeStepS))
		var bc BodyController
		if bc.Target(body, body.Entity) {
			bc.findBodySector()
		}
	}
}
