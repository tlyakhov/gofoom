package state

import (
	"reflect"
	"sync"
	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
)

type EditorTool int

const (
	ToolSelect EditorTool = iota
	ToolSplitSegment
	ToolSplitSector
	ToolAddSector
	ToolAddBody
	ToolAlignGrid
)

type Edit struct {
	MapView
	DB   *concepts.EntityComponentDB
	Lock sync.RWMutex

	// Map view positions in world/screen space.
	Mouse          concepts.Vector2 // Screen
	MouseDown      concepts.Vector2 // Screen
	MouseWorld     concepts.Vector2
	MouseDownWorld concepts.Vector2
	MousePressed   bool

	SelectedObjects        []any
	HoveringObjects        []any
	SearchTerms            string
	SelectedTransformables []any

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
	SetMapCursor(cursor desktop.Cursor)
	UpdateTitle()
	Load(filename string)
	ActionFinished(canceled, refreshProperties, autoPortal bool)
	NewAction(a IAction)
	ActTool()
	SwitchTool(tool EditorTool)
	UndoCurrent()
	RedoCurrent()
	SelectObjects(objects []any, updateTree bool)
	Selecting() bool
	SelectionBox() (v1 *concepts.Vector2, v2 *concepts.Vector2)
	Alert(text string)
	SetDialogLocation(dlg *dialog.FileDialog, target string)
}

func IndexOf(s []any, obj any) int {
	if er, ok := obj.(*concepts.EntityRef); ok {
		for i, e := range s {
			if er2, ok2 := e.(*concepts.EntityRef); ok2 && er2.Entity == er.Entity {
				return i
			}
		}
		return -1
	}
	for i, e := range s {
		if obj == e && reflect.TypeOf(obj) == reflect.TypeOf(e) {
			return i
		}
	}
	return -1
}
