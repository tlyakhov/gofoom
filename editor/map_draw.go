package main

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/actions"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

func TransformContext(cr *cairo.Context) {
	t := editor.Pos.Mul(-editor.Scale).Add(editor.Size.Mul(0.5))
	cr.Translate(t.X, t.Y)
	cr.Scale(editor.Scale, editor.Scale)
}

func DrawMap(da *gtk.DrawingArea, cr *cairo.Context) {
	w := da.GetAllocatedWidth()
	h := da.GetAllocatedHeight()
	editor.Size = concepts.Vector2{X: float64(w), Y: float64(h)}

	editor.MapViewGrid.Draw(&editor.Edit, cr)
	TransformContext(cr)

	for _, sector := range editor.World.Sectors {
		DrawSector(cr, sector)
	}

	switch editor.CurrentAction.(type) {
	case *actions.Select:
		if editor.MousePressed {
			v1, v2 := editor.SelectionBox()
			cr.Rectangle(v1.X, v1.Y, v2.X-v1.X, v2.Y-v1.Y)
			cr.SetSourceRGBA(0.2, 0.2, 1.0, 0.3)
			cr.Fill()
			cr.SetSourceRGBA(0.67, 0.67, 1.0, 0.3)
			cr.Stroke()
		}
	case *actions.AddSector:
		gridMouse := editor.WorldGrid(editor.MouseWorld)
		cr.SetSourceRGB(ColorSelectionPrimary.X, ColorSelectionPrimary.Y, ColorSelectionPrimary.Z)
		DrawHandle(cr, gridMouse)
	case *actions.SplitSector, *actions.SplitSegment, *actions.AlignGrid:
		gridMouse := editor.WorldGrid(editor.MouseWorld)
		gridMouseDown := editor.WorldGrid(editor.MouseDownWorld)
		cr.SetSourceRGB(ColorSelectionPrimary.X, ColorSelectionPrimary.Y, ColorSelectionPrimary.Z)
		DrawHandle(cr, gridMouse)
		if editor.MousePressed {
			cr.NewPath()
			cr.MoveTo(gridMouseDown.X, gridMouseDown.Y)
			cr.LineTo(gridMouse.X, gridMouse.Y)
			cr.ClosePath()
			cr.Stroke()
			DrawHandle(cr, gridMouseDown)
		}

	}

	//cr.ShowText(fmt.Sprintf("%v, %v", Mouse.X, Mouse.Y))
}
