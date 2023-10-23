package main

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/actions"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

func TransformContext(cr *cairo.Context) {
	t := editor.Pos.Mul(-editor.Scale).Add(editor.Size.Mul(0.5))
	cr.Translate(t[0], t[1])
	cr.Scale(editor.Scale, editor.Scale)
}

func DrawMap(da *gtk.DrawingArea, cr *cairo.Context) {
	w := da.GetAllocatedWidth()
	h := da.GetAllocatedHeight()
	editor.Size = concepts.Vector2{float64(w), float64(h)}

	editor.MapViewGrid.Draw(&editor.Edit, cr)
	TransformContext(cr)

	for _, isector := range editor.DB.All(core.SectorComponentIndex) {
		DrawSector(cr, isector.(*core.Sector))
	}

	switch editor.CurrentAction.(type) {
	case *actions.Select:
		if editor.MousePressed {
			v1, v2 := editor.SelectionBox()
			cr.Rectangle(v1[0], v1[1], v2[0]-v1[0], v2[1]-v1[1])
			cr.SetSourceRGBA(0.2, 0.2, 1.0, 0.3)
			cr.Fill()
			cr.SetSourceRGBA(0.67, 0.67, 1.0, 0.3)
			cr.Stroke()
		}
	case *actions.AddSector:
		gridMouse := editor.WorldGrid(&editor.MouseWorld)
		cr.SetSourceRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
		DrawHandle(cr, gridMouse)
	case *actions.SplitSector, *actions.SplitSegment, *actions.AlignGrid:
		gridMouse := editor.WorldGrid(&editor.MouseWorld)
		gridMouseDown := editor.WorldGrid(&editor.MouseDownWorld)
		cr.SetSourceRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
		DrawHandle(cr, gridMouse)
		if editor.MousePressed {
			cr.NewPath()
			cr.MoveTo(gridMouseDown[0], gridMouseDown[1])
			cr.LineTo(gridMouse[0], gridMouse[1])
			cr.ClosePath()
			cr.Stroke()
			DrawHandle(cr, gridMouseDown)
		}

	}

	//cr.ShowText(fmt.Sprintf("%v, %v", Mouse[0], Mouse[1]))
}
