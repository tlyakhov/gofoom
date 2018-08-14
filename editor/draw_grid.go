package main

import (
	"fmt"

	"github.com/gotk3/gotk3/cairo"
	"github.com/tlyakhov/gofoom/concepts"
)

func DrawGrid(cr *cairo.Context) {
	if !editor.Grid.Visible || editor.Scale*GridSize < 5.0 {
		cr.SetSourceRGB(0, 0, 0)
		cr.Paint()
		return
	}

	if editor.Grid.Prev != editor.MapViewState {
		fmt.Println("change")
		RefreshGrid(cr)
		editor.Grid.Prev = editor.MapViewState
	}

	//editor.Grid.Surface.MarkDirty()
	cr.SetSourceSurface(editor.Grid.Surface, 0, 0)
	cr.Paint()
}

func RefreshGrid(cr *cairo.Context) {
	//target := cr.GetTarget()
	//editor.Grid.Surface = target.CreateSimilar(cairo.CONTENT_COLOR_ALPHA, target.GetWidth(), target.GetHeight())
	cr.PushGroup()
	//gcr := cairo.Create(editor.Grid.Surface)

	cr.SetSourceRGB(0, 0, 0)
	cr.Paint()
	TransformContext(cr)
	cr.SetSourceRGB(0.5, 0.5, 0.5)

	start := editor.ScreenToWorld(concepts.Vector2{}).Mul(1.0 / GridSize).Floor().Mul(GridSize)
	end := editor.ScreenToWorld(editor.MapViewSize).Mul(1.0 / GridSize).Add(concepts.Vector2{1, 1}).Floor().Mul(GridSize)
	for x := start.X; x < end.X; x += GridSize {
		for y := start.Y; y < end.Y; y += GridSize {
			cr.Rectangle(x, y, 1, 1)
			cr.Fill()
		}
	}
	editor.Grid.Surface = cr.GetGroupTarget()
	cr.PopGroupToSource()

	//editor.Grid.Surface.Flush()
	//editor.Grid.Surface.MarkDirty()

}
