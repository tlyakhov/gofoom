package main

import (
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/cairo"
)

type MapViewGrid struct {
	Current *state.MapView
	Prev    *state.MapView
	Visible bool
	Surface *cairo.Surface
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

	r := &concepts.Vector3{}
	g.WorldGrid(&concepts.Vector2{p[0], p[1]}).To3D(r)
	r[2] = p[2]
	return r
}

func (g *MapViewGrid) Draw(e *state.Edit, cr *cairo.Context) {
	if !g.Visible || e.Scale*g.Current.Step < 5.0 {
		cr.SetSourceRGB(0, 0, 0)
		cr.Paint()
		return
	}

	if g.Prev == nil || *g.Prev != *g.Current {
		g.Refresh(e, cr)
		mv := *g.Current
		g.Prev = &mv
	}

	cr.Save()
	cr.IdentityMatrix()
	cr.SetSourceSurface(editor.MapViewGrid.Surface, 0, 0)
	cr.Paint()
	cr.Restore()
}

func (g *MapViewGrid) Refresh(e *state.Edit, cr *cairo.Context) {
	cr.PushGroup()
	cr.SetSourceRGB(0, 0, 0)
	cr.Paint()
	TransformContext(cr)
	cr.SetSourceRGB(0.5, 0.5, 0.5)

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
	start := editor.ScreenToWorld(&concepts.Vector2{})
	end := editor.ScreenToWorld(&editor.Size)
	qstart := g.WorldGrid(start)
	qend := g.WorldGrid(end)
	delta := qend.Sub(qstart).Mul(1.0 / g.Current.Step).Floor()
	qstart = qstart.Sub(right.Mul(delta[0])).Sub(down.Mul(delta[1]))
	// qend = qend.Add(right.Mul(delta[0])).Add(down.Mul(delta[1]))

	d := 3.0 / (g.Current.Scale + 1)

	for x := 0.0; x < delta[0]*3; x++ {
		for y := 0.0; y < delta[1]*3; y++ {
			pos := qstart.Add(right.Mul(x)).Add(down.Mul(y))
			if pos[0] < start[0] || pos[0] > end[0] || pos[1] < start[1] || pos[1] > end[1] {
				continue
			}
			cr.Rectangle(pos[0]-d*0.5, pos[1]-d*0.5, d, d)
			cr.Fill()
		}
	}
	editor.MapViewGrid.Surface = cr.GetGroupTarget()
	cr.PopGroupToSource()
}
