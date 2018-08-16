package main

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
	"github.com/tlyakhov/gofoom/concepts"
)

func TransformContext(cr *cairo.Context) {
	t := editor.Pos.Mul(-editor.Scale).Add(editor.MapViewSize.Mul(0.5))
	cr.Translate(t.X, t.Y)
	cr.Scale(editor.Scale, editor.Scale)
}

func DrawMap(da *gtk.DrawingArea, cr *cairo.Context) {
	w := da.GetAllocatedWidth()
	h := da.GetAllocatedHeight()
	editor.MapViewSize = concepts.Vector2{float64(w), float64(h)}

	DrawMapGrid(cr)
	TransformContext(cr)

	for _, sector := range editor.GameMap.Sectors {
		DrawSector(cr, sector)
	}

	if editor.Selecting() {
		v1, v2 := editor.SelectionBox()
		cr.Rectangle(v1.X, v1.Y, v2.X-v1.X, v2.Y-v1.Y)
		cr.SetSourceRGBA(0.2, 0.2, 1.0, 0.3)
		cr.Fill()
		cr.SetSourceRGBA(0.67, 0.67, 1.0, 0.3)
		cr.Stroke()
	}

	//cr.ShowText(fmt.Sprintf("%v, %v", Mouse.X, Mouse.Y))
}
