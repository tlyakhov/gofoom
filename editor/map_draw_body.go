// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"math"

	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"

	"tlyakhov/gofoom/components/behaviors"
)

func (mw *MapWidget) DrawBodyAngle(e *core.Body) {
	astart := (e.Angle.Render - editor.Renderer.FOV/2) * math.Pi / 180.0
	aend := (e.Angle.Render + editor.Renderer.FOV/2) * math.Pi / 180.0
	diameter := e.Size.Render[0]
	mw.Context.SetLineWidth(2)
	mw.Context.NewSubPath()
	mw.Context.MoveTo(e.Pos.Render[0], e.Pos.Render[1])
	mw.Context.LineTo(e.Pos.Render[0]+math.Cos(astart)*diameter, e.Pos.Render[1]+math.Sin(astart)*diameter)
	mw.Context.DrawArc(e.Pos.Render[0], e.Pos.Render[1], diameter, astart, aend)
	mw.Context.MoveTo(e.Pos.Render[0], e.Pos.Render[1])
	mw.Context.LineTo(e.Pos.Render[0]+math.Cos(aend)*diameter, e.Pos.Render[1]+math.Sin(aend)*diameter)
	mw.Context.ClosePath()
	mw.Context.Stroke()
}

func (mw *MapWidget) DrawBody(body *core.Body) {
	start := editor.ScreenToWorld(new(concepts.Vector2))
	end := editor.ScreenToWorld(&editor.Size)
	if body.Pos.Render[0] < start[0] || body.Pos.Render[0] > end[0] ||
		body.Pos.Render[1] < start[1] || body.Pos.Render[1] > end[1] {
		return
	}

	mw.Context.SetRGB(0.6, 0.6, 0.6)
	if player := character.GetPlayer(body.Entity); player != nil {
		// Let's get fancy:
		mw.Context.SetRGB(0.6, 0.6, 0.6)
		mw.Context.SetLineWidth(1)
		mw.Context.NewSubPath()
		mw.Context.DrawArc(body.Pos.Render[0], body.Pos.Render[1], body.Size.Render[0]*0.25, 0, math.Pi*2)
		mw.Context.ClosePath()
		mw.Context.Stroke()
	} else if light := core.GetLight(body.Entity); light != nil {
		mw.Context.SetRGB(light.Diffuse[0], light.Diffuse[1], light.Diffuse[2])
	} // Sprite...

	hovering := editor.HoveringObjects.Contains(selection.SelectableFromBody(body))
	selected := editor.SelectedObjects.Contains(selection.SelectableFromBody(body))

	if selected || hovering {
		img := editor.EntityImage(body.Entity)
		mw.Context.Push()
		mw.Context.Translate(body.Pos.Render[0]+2*body.Size.Render[0], body.Pos.Render[1])
		mw.Context.Scale(2*body.Size.Render[0]/64, 2*body.Size.Render[0]/64)
		mw.Context.DrawImageAnchored(img, 0, 0, 0.5, 0.5)
		mw.Context.Pop()
	}

	if selected {
		mw.Context.SetStrokeStyle(PatternSelectionPrimary)
	} else if hovering {
		mw.Context.SetStrokeStyle(PatternSelectionSecondary)
	}

	mw.Context.SetLineWidth(1)
	mw.Context.NewSubPath()
	mw.Context.DrawArc(body.Pos.Render[0], body.Pos.Render[1], body.Size.Render[0]*0.5, 0, math.Pi*2)
	mw.Context.ClosePath()
	mw.Context.Stroke()
	mw.Context.SetRGB(0.33, 0.33, 0.33)
	mw.DrawBodyAngle(body)

	if prox := behaviors.GetProximity(body.Entity); prox != nil {
		mw.Context.SetRGB(0, 0.33, 0.33)
		mw.Context.NewSubPath()
		mw.Context.DrawArc(body.Pos.Render[0], body.Pos.Render[1], prox.Range, 0, math.Pi*2)
		mw.Context.ClosePath()
		mw.Context.Stroke()
	}

	if editor.ComponentNamesVisible {
		text := body.Format()
		mw.Context.Push()
		mw.Context.SetRGB(0.3, 0.3, 0.5)
		mw.Context.DrawStringAnchored(text, body.Pos.Render[0], body.Pos.Render[1], 0.5, 0.5)
		mw.Context.Pop()
	}
}
