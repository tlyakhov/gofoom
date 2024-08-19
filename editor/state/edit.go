// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"image"
	"sync"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"

	"fyne.io/fyne/v2"
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
	ToolAddInternalSegment
	ToolAlignGrid
)

type Edit struct {
	MapView
	ECS  *ecs.ECS
	Lock sync.Mutex

	// Map view positions in world/screen space.
	Mouse          concepts.Vector2 // Screen
	MouseDown      concepts.Vector2 // Screen
	MouseWorld     concepts.Vector2
	MouseDownWorld concepts.Vector2
	MousePressed   bool
	Dragging       bool

	SelectedObjects        *core.Selection
	HoveringObjects        *core.Selection
	SearchTerms            string
	SelectedTransformables []any

	Tool          EditorTool
	OpenFile      string
	Modified      bool
	CurrentAction Actionable
	UndoHistory   []Actionable
	RedoHistory   []Actionable
	KeysDown      concepts.Set[fyne.KeyName]

	// Map view filters
	BodiesVisible         bool
	SectorTypesVisible    bool
	ComponentNamesVisible bool
}

type IEditor interface {
	fyne.Clipboard
	State() *Edit
	ScreenToWorld(p *concepts.Vector2) *concepts.Vector2
	WorldToScreen(p *concepts.Vector2) *concepts.Vector2
	WorldGrid(p *concepts.Vector2) *concepts.Vector2
	WorldGrid3D(p *concepts.Vector3) *concepts.Vector3
	SetMapCursor(cursor desktop.Cursor)
	UpdateTitle()
	Load(filename string)
	ActionFinished(canceled, refreshProperties, autoPortal bool)
	NewAction(a Actionable)
	ActTool()
	SwitchTool(tool EditorTool)
	UndoCurrent()
	RedoCurrent()
	SelectObjects(updateEntityList bool, s ...*core.Selectable)
	SetSelection(updateEntityList bool, s *core.Selection)
	Selecting() bool
	SelectionBox() (v1 *concepts.Vector2, v2 *concepts.Vector2)
	Alert(text string)
	SetDialogLocation(dlg *dialog.FileDialog, target string)
	EntityImage(entity ecs.Entity, sector bool) image.Image
}
