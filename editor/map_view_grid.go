// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"image"
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"github.com/fogleman/gg"
)

type MapViewGrid struct {
	Current     *state.MapView
	Prev        *state.MapView
	Visible     bool
	GridContext *gg.Context
}

func (g *MapViewGrid) WorldGrid(p *concepts.Vector2) *concepts.Vector2 {
	if !g.Visible {
		return p
	}

	down := g.Current.GridB.Sub(&g.Current.GridA).Norm()
	right := &concepts.Vector2{down[1], -down[0]}
	p = p.Sub(&g.Current.GridA)
	p[0], p[1] = p[0]*down[1]-p[1]*down[0], -p[0]*right[1]+p[1]*right[0]
	p[0], p[1] = math.Round(p[0]/g.Current.Step)*g.Current.Step, math.Round(p[1]/g.Current.Step)*g.Current.Step
	p[0], p[1] = p[0]*right[0]+p[1]*down[0], p[0]*right[1]+p[1]*down[1]
	p = p.AddSelf(&g.Current.GridA)
	return p
}

func (g *MapViewGrid) WorldGrid3D(p *concepts.Vector3) *concepts.Vector3 {
	if !g.Visible {
		return p
	}

	r := new(concepts.Vector3)
	g.WorldGrid(&concepts.Vector2{p[0], p[1]}).To3D(r)
	r[2] = p[2]
	return r
}

func (g *MapViewGrid) Draw(e *state.Edit) {
	if !g.Visible || e.Scale*g.Current.Step < 5.0 {
		g.Clear()
		return
	}

	if g.Prev == nil || *g.Prev != *g.Current {
		g.Refresh(e)
		mv := *g.Current
		g.Prev = &mv
	}
}

func (g *MapViewGrid) pixels() []uint8 {
	return g.GridContext.Image().(*image.RGBA).Pix
}

func (g *MapViewGrid) size() (int, int) {
	img := g.GridContext.Image().(*image.RGBA)
	return img.Rect.Dx(), img.Rect.Dy()
}

func (g *MapViewGrid) Clear() {
	pix := g.pixels()

	for i := 0; i < len(pix)/4; i++ {
		pix[i*4+0] = 0
		pix[i*4+1] = 0
		pix[i*4+2] = 0
		pix[i*4+3] = 0xFF
	}
}

func (g *MapViewGrid) Refresh(e *state.Edit) {
	down := g.Current.GridB.Sub(&g.Current.GridA).Norm().Mul(g.Current.Step)
	right := &concepts.Vector2{down[1], -down[0]}
	if math.Abs(down[0]) > math.Abs(right[0]) {
		down, right = right, down
	}
	if down[1] < 0 {
		down[0] = -down[0]
		down[1] = -down[1]
	}
	if right[0] < 0 {
		right[0] = -right[0]
		right[1] = -right[1]
	}
	start := editor.ScreenToWorld(new(concepts.Vector2))
	end := editor.ScreenToWorld(&editor.Size)
	qstart := g.WorldGrid(start)
	qend := g.WorldGrid(end)
	delta := qend.Sub(qstart).Mul(1.0 / g.Current.Step).Floor()
	qstart = qstart.Sub(right.Mul(delta[0])).Sub(down.Mul(delta[1]))
	// qend = qend.Add(right.Mul(delta[0])).Add(down.Mul(delta[1]))

	//	d := 3.0 / (g.Current.Scale + 1)
	var pos concepts.Vector2
	g.Clear()
	pix := g.pixels()
	w, h := g.size()
	for x := 0.0; x < delta[0]*3; x++ {
		for y := 0.0; y < delta[1]*3; y++ {
			pos[0] = qstart[0] + right[0]*x + down[0]*y
			pos[1] = qstart[1] + right[1]*x + down[1]*y
			/*if pos[0] < start[0] || pos[0] > end[0] || pos[1] < start[1] || pos[1] > end[1] {
				continue
			}*/
			v := editor.WorldToScreen(&pos)
			ix := int(v[0])
			iy := int(v[1])
			if ix >= 0 && ix < w &&
				iy >= 0 && iy < h {
				index := ix + iy*w
				pix[index*4+0] = 0x80
				pix[index*4+1] = 0x80
				pix[index*4+2] = 0x80
				pix[index*4+3] = 0xFF
			}
		}
	}
}
