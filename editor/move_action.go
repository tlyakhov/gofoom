package main

import (
	"math"

	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/logic/provide"

	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/core"
)

type MoveAction struct {
	*Editor
	Selected []concepts.ISerializable
	Original []concepts.Vector3
	Delta    concepts.Vector2
}

func (a *MoveAction) OnMouseDown(button *gdk.EventButton) {
	a.Selected = make([]concepts.ISerializable, len(a.SelectedObjects))
	a.Original = make([]concepts.Vector3, len(a.SelectedObjects))
	copy(a.Selected, a.SelectedObjects)

	for i, obj := range a.Selected {
		switch target := obj.(type) {
		case *MapPoint:
			a.Original[i] = target.A.To3D()
		case core.AbstractEntity:
			a.Original[i] = target.Physical().Pos
		}
	}
}

func (a *MoveAction) OnMouseMove() {
	a.Delta = a.MouseWorld.Sub(a.MouseDownWorld)
	a.Act()
}

func (a *MoveAction) OnMouseUp() {
	a.Editor.SelectObjects(a.SelectedObjects) // Updates properties.
	a.ActionFinished()
}
func (a *MoveAction) Act() {
	for i, obj := range a.Selected {
		switch target := obj.(type) {
		case *MapPoint:
			target.A = a.Original[i].To2D().Add(a.Delta)
			if a.Grid.Visible {
				target.A.X = math.Round(target.A.X/GridSize) * GridSize
				target.A.Y = math.Round(target.A.Y/GridSize) * GridSize
			}
			provide.Passer.For(target.Sector).Recalculate()
		case core.AbstractEntity:
			if target == a.GameMap.Player {
				// Otherwise weird things happen...
				continue
			}
			target.Physical().Pos = a.Original[i].Add(a.Delta.To3D())
			if a.Grid.Visible {
				target.Physical().Pos.X = math.Round(target.Physical().Pos.X/GridSize) * GridSize
				target.Physical().Pos.Y = math.Round(target.Physical().Pos.Y/GridSize) * GridSize
			}
			if c, ok := provide.Collider.For(target); ok {
				c.Collide()
			}
			if _, ok := target.(*entities.Light); ok {
				a.GameMap.ClearLightmaps()
				a.GameMap.Recalculate()
			}
		}
	}
}
func (a *MoveAction) Cancel() {}
func (a *MoveAction) Frame()  {}

func (a *MoveAction) Undo() {
	for i, obj := range a.Selected {
		switch target := obj.(type) {
		case *MapPoint:
			target.A = a.Original[i].To2D()
			provide.Passer.For(target.Sector).Recalculate()
		case core.AbstractEntity:
			if target == a.GameMap.Player {
				// Otherwise weird things happen...
				continue
			}
			target.Physical().Pos = a.Original[i]
			if c, ok := provide.Collider.For(target); ok {
				c.Collide()
			}
			if _, ok := target.(*entities.Light); ok {
				a.GameMap.ClearLightmaps()
			}
		}
	}
}
func (a *MoveAction) Redo() {
	a.Act()
}
