package main

import (
	"fmt"
	"math"

	"github.com/gotk3/gotk3/cairo"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/editor/state"
)

type MapViewGrid struct {
	Current *state.MapView
	Prev    *state.MapView
	Visible bool
	Surface *cairo.Surface
}

func (g *MapViewGrid) WorldGrid(p concepts.Vector2) concepts.Vector2 {
	if !g.Visible {
		return p
	}

	down := g.Current.GridB.Sub(g.Current.GridA).Norm()
	right := concepts.Vector2{down.Y, -down.X}
	p = p.Sub(g.Current.GridA)
	p = concepts.Vector2{p.X*down.Y - p.Y*down.X, -p.X*right.Y + p.Y*right.X}
	p.X = math.Round(p.X/g.Current.Step) * g.Current.Step
	p.Y = math.Round(p.Y/g.Current.Step) * g.Current.Step
	p = concepts.Vector2{p.X*right.X + p.Y*down.X, p.X*right.Y + p.Y*down.Y}
	p = p.Add(g.Current.GridA)
	return p
}

func (g *MapViewGrid) WorldGrid3D(p concepts.Vector3) concepts.Vector3 {
	if !g.Visible {
		return p
	}

	r := g.WorldGrid(p.To2D()).To3D()
	r.Z = p.Z
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

	down := g.Current.GridB.Sub(g.Current.GridA).Norm().Mul(g.Current.Step)
	right := concepts.Vector2{down.Y, -down.X}
	if math.Abs(down.X) > math.Abs(right.X) {
		down, right = right, down
	}
	if down.Y < 0 {
		down.X = -down.X
		down.Y = -down.Y
	}
	if right.X < 0 {
		right.X = -right.X
		right.Y = -right.Y
	}
	start := editor.ScreenToWorld(concepts.Vector2{})
	end := editor.ScreenToWorld(editor.Size)
	qstart := g.WorldGrid(start)
	qend := g.WorldGrid(end)
	delta := qend.Sub(qstart).Mul(1.0 / g.Current.Step).Floor()
	qstart = qstart.Sub(right.Mul(delta.X)).Sub(down.Mul(delta.Y))
	qend = qend.Add(right.Mul(delta.X)).Add(down.Mul(delta.Y))

	d := 3.0 / (g.Current.Scale + 1)
	fmt.Printf("%v\n", d)

	for x := 0.0; x < delta.X*3; x++ {
		for y := 0.0; y < delta.Y*3; y++ {
			pos := qstart.Add(right.Mul(x)).Add(down.Mul(y))
			if pos.X < start.X || pos.X > end.X || pos.Y < start.Y || pos.Y > end.Y {
				continue
			}
			cr.Rectangle(pos.X-d*0.5, pos.Y-d*0.5, d, d)
			cr.Fill()
		}
	}
	editor.MapViewGrid.Surface = cr.GetGroupTarget()
	cr.PopGroupToSource()
}
