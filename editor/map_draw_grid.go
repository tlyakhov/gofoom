package main

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/tlyakhov/gofoom/concepts"
)

func DrawMapGrid(cr *cairo.Context) {
	if !editor.MapViewGrid.Visible || editor.Scale*GridSize < 5.0 {
		cr.SetSourceRGB(0, 0, 0)
		cr.Paint()
		return
	}

	if editor.MapViewGrid.Prev != editor.MapView {
		RefreshMapGrid(cr)
		editor.MapViewGrid.Prev = editor.MapView
	}

	cr.Save()
	cr.IdentityMatrix()
	cr.SetSourceSurface(editor.MapViewGrid.Surface, 0, 0)
	cr.Paint()
	cr.Restore()
}

func RefreshMapGrid(cr *cairo.Context) {
	cr.PushGroup()
	cr.SetSourceRGB(0, 0, 0)
	cr.Paint()
	TransformContext(cr)
	cr.SetSourceRGB(0.5, 0.5, 0.5)

	start := editor.ScreenToWorld(concepts.Vector2{}).Mul(1.0 / GridSize).Floor().Mul(GridSize)
	end := editor.ScreenToWorld(editor.Size).Mul(1.0 / GridSize).Add(concepts.Vector2{1, 1}).Floor().Mul(GridSize)
	for x := start.X; x < end.X; x += GridSize {
		for y := start.Y; y < end.Y; y += GridSize {
			cr.Rectangle(x, y, 1, 1)
			cr.Fill()
		}
	}
	editor.MapViewGrid.Surface = cr.GetGroupTarget()
	cr.PopGroupToSource()
}
