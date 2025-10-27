// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"image"
	"sync"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/containers"
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

type EditorState struct {
	MapView
	Lock          sync.Mutex
	GameInputLock sync.Mutex

	// Map view positions in world/screen space.
	Mouse          concepts.Vector2 // Screen
	MouseDown      concepts.Vector2 // Screen
	MouseWorld     concepts.Vector2
	MouseDownWorld concepts.Vector2
	MousePressed   bool
	Dragging       bool

	SelectedObjects        *selection.Selection
	HoveringObjects        *selection.Selection
	SearchQuery            string
	SelectedTransformables []any

	Tool          EditorTool
	OpenFile      string
	Modified      bool
	CurrentAction Actionable
	/*
		TODO: Re-implement undo/redo by serializing/deserializing the entire world
		state instead. The current system is too brittle and requires a ton of work
		for every individual action to store and apply diffs.

		Tricky aspects:
		1. SourceFiles. Restoring in-memory snapshots will involve reloading source
		   files every undo/redo. That seems expensive and unnecessary. Can we avoid
		   doing this?
		2. Dynamics/Simulation details. Probably nothing will break, but diffs will
		   have noise from this.
	*/
	UndoHistory []Actionable
	RedoHistory []Actionable
	KeysDown    containers.Set[fyne.KeyName]

	// Map view filters
	BodiesVisible         bool
	SectorTypesVisible    bool
	ComponentNamesVisible bool
}

type IEditor interface {
	fyne.Clipboard
	State() *EditorState
	ScreenToWorld(p *concepts.Vector2) *concepts.Vector2
	WorldToScreen(p *concepts.Vector2) *concepts.Vector2
	WorldGrid(p *concepts.Vector2) *concepts.Vector2
	WorldGrid3D(p *concepts.Vector3) *concepts.Vector3
	SetMapCursor(cursor desktop.Cursor)
	UpdateTitle()
	Load(filename string)
	ActionFinished(canceled, refreshProperties, autoPortal bool)
	Act(a Actionable)
	UseTool()
	SwitchTool(tool EditorTool)
	UndoCurrent()
	RedoCurrent()
	SelectObjects(updateEntityList bool, s ...*selection.Selectable)
	SetSelection(updateEntityList bool, s *selection.Selection)
	Selecting() bool
	SelectionBox() (v1 *concepts.Vector2, v2 *concepts.Vector2)
	Alert(text string)
	SetDialogLocation(dlg *dialog.FileDialog, target string)
	EntityImage(entity ecs.Entity) image.Image
	FlushEntityImage(entity ecs.Entity)
}
