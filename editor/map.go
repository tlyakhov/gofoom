package main

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
	"github.com/tlyakhov/gofoom/concepts"
)

func DrawMap(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.PushGroup()
	w := da.GetAllocatedWidth()
	h := da.GetAllocatedHeight()
	editor.MapViewSize = concepts.Vector2{float64(w), float64(h)}

	cr.IdentityMatrix()
	cr.SetSourceRGB(0, 0, 0)
	cr.Paint()
	t := editor.Pos.Mul(-editor.Scale).Add(editor.MapViewSize.Mul(0.5))
	cr.Translate(t.X, t.Y)
	cr.Scale(editor.Scale, editor.Scale)

	if editor.GridVisible && editor.Scale*GridSize > 5.0 {
		start := editor.ScreenToWorld(concepts.Vector2{}).Mul(1.0 / GridSize).Floor().Mul(GridSize)
		end := editor.ScreenToWorld(editor.MapViewSize).Mul(1.0 / GridSize).Add(concepts.Vector2{1, 1}).Floor().Mul(GridSize)

		cr.SetSourceRGB(0.5, 0.5, 0.5)
		for x := start.X; x < end.X; x += GridSize {
			for y := start.Y; y < end.Y; y += GridSize {
				cr.Rectangle(x, y, 1, 1)
				cr.Fill()
			}
		}
	}

	// Hovering

	editor.HoveringObjects = make(map[string]concepts.ISerializable)

	v1 := editor.MouseWorld
	v2 := editor.MouseDownWorld

	if editor.MousePressed && v2.X < v1.X {
		tmp := v1.X
		v1.X = v2.X
		v2.X = tmp
	}
	if editor.MousePressed && v2.Y < v1.Y {
		tmp := v1.Y
		v1.Y = v2.Y
		v2.Y = tmp
	}
	for _, sector := range editor.GameMap.Sectors {
		phys := sector.Physical()

		for _, segment := range phys.Segments {
			if editor.CurrentAction == nil {
				if editor.Mouse.Sub(editor.WorldToScreen(segment.A)).Length() < SegmentSelectionEpsilon {
					editor.HoveringObjects[segment.ID] = segment
				}
			}
		}

	}

	// Drawing

	for _, sector := range editor.GameMap.Sectors {
		DrawSector(cr, sector)
	}

	//cr.ShowText(fmt.Sprintf("%v, %v", Mouse.X, Mouse.Y))
	cr.PopGroupToSource()
	cr.Paint()
}
