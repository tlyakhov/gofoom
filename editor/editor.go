package main

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/logic"
)

type Editor struct {
	State              string
	Scale              float64
	Pos                concepts.Vector2
	Mouse              concepts.Vector2
	MouseDown          concepts.Vector2
	MouseWorld         concepts.Vector2
	MouseDownWorld     concepts.Vector2
	GameMap            *logic.MapService
	MousePressed       bool
	GridVisible        bool
	EntitiesVisible    bool
	SectorTypesVisible bool
	MapViewSize        concepts.Vector2
	HoveringObjects    map[string]concepts.ISerializable
	SelectedObjects    map[string]concepts.ISerializable
	CurrentAction      AbstractAction
	UndoHistory        []AbstractAction
	RedoHistory        []AbstractAction
}

func NewEditor() *Editor {
	return &Editor{
		Scale:              1.0,
		GridVisible:        true,
		EntitiesVisible:    true,
		SectorTypesVisible: true,
	}
}

func (e *Editor) ScreenToWorld(p concepts.Vector2) concepts.Vector2 {
	return p.Sub(e.MapViewSize.Mul(0.5)).Mul(1.0 / e.Scale).Add(e.Pos)
}

func (e *Editor) WorldToScreen(p concepts.Vector2) concepts.Vector2 {
	return p.Sub(e.Pos).Mul(e.Scale).Add(e.MapViewSize.Mul(0.5))
}

func (e *Editor) ActionFinished() {
	// Set Cursor
	e.State = "Idle"
	e.CurrentAction = nil
	e.ActTool()
}

func (e *Editor) NewAction(a AbstractAction) {
	e.CurrentAction = a
	e.UndoHistory = append(e.UndoHistory, a)
	if len(e.UndoHistory) > 100 {
		e.UndoHistory = e.UndoHistory[(len(e.UndoHistory) - 100):]
	}
	e.RedoHistory = []AbstractAction{}
}

func (e *Editor) ActTool() {

}

func (e *Editor) Undo() {
	index := len(e.UndoHistory) - 1
	if index < 0 {
		return
	}
	a := e.UndoHistory[index]
	e.UndoHistory = e.UndoHistory[:index]
	if a == nil {
		return
	}
	a.Undo()
	e.RedoHistory = append(e.RedoHistory, a)
}

func (e *Editor) Redo() {
	index := len(e.RedoHistory) - 1
	if index < 0 {
		return
	}
	a := e.RedoHistory[index]
	e.RedoHistory = e.RedoHistory[:index]
	if a == nil {
		return
	}
	a.Redo()
	e.UndoHistory = append(e.UndoHistory, a)
}

func (e *Editor) SelectObjects(objects map[string]concepts.ISerializable) {
	if objects == nil || len(objects) == 0 {
		objects = make(map[string]concepts.ISerializable)
		objects[e.GameMap.ID] = e.GameMap
	}

	e.SelectedObjects = objects
	// e.RefreshPropertyGrid
}
