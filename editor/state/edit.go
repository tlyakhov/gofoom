package state

import (
	"reflect"
	"tlyakhov/gofoom/concepts"
)

type EditorTool int

const (
	ToolSelect EditorTool = iota
	ToolSplitSegment
	ToolSplitSector
	ToolAddSector
	ToolAddEntity
	ToolAlignGrid
)

type Edit struct {
	MapView
	DB *concepts.EntityComponentDB

	// Map view positions in world/screen space.
	Mouse          concepts.Vector2 // Screen
	MouseDown      concepts.Vector2 // Screen
	MouseWorld     concepts.Vector2
	MouseDownWorld concepts.Vector2
	MousePressed   bool

	SelectedObjects []any
	HoveringObjects []any
	Filter          string

	Tool          EditorTool
	OpenFile      string
	Modified      bool
	CurrentAction IAction
	UndoHistory   []IAction
	RedoHistory   []IAction
}

type IEditor interface {
	State() *Edit
	ScreenToWorld(p *concepts.Vector2) *concepts.Vector2
	WorldToScreen(p *concepts.Vector2) *concepts.Vector2
	WorldGrid(p *concepts.Vector2) *concepts.Vector2
	WorldGrid3D(p *concepts.Vector3) *concepts.Vector3
	SetMapCursor(name string)
	UpdateTitle()
	Load(filename string)
	ActionFinished(canceled bool)
	NewAction(a IAction)
	ActTool()
	SwitchTool(tool EditorTool)
	UndoCurrent()
	RedoCurrent()
	SelectObjects(objects []any)
	Selecting() bool
	SelectionBox() (v1 *concepts.Vector2, v2 *concepts.Vector2)
	Alert(text string)
}

func IndexOf(s []any, obj any) int {
	//v := reflect.ValueOf(obj)
	for i, e := range s {
		if obj == e && reflect.TypeOf(obj) == reflect.TypeOf(e) {
			return i
		}
	}
	return -1
}
