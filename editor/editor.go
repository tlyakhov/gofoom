package main

import (
	"fmt"
	"strconv"
	"unsafe"

	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/editor/actions"

	"github.com/gotk3/gotk3/gdk"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/sectors"
	"tlyakhov/gofoom/editor/properties"
	"tlyakhov/gofoom/editor/state"
	"tlyakhov/gofoom/render"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/controllers"
)

type EditorWidgets struct {
	App               *gtk.Application
	Window            *gtk.ApplicationWindow
	GameArea          *gtk.DrawingArea
	MapArea           *gtk.DrawingArea
	EntitySearchBar   *gtk.SearchBar
	EntitySearchEntry *gtk.SearchEntry
	StatusBar         *gtk.Label
}

type Editor struct {
	state.Edit
	// What we're editing.

	MapViewGrid
	EditorWidgets
	properties.Grid
	properties.EntityTree

	// Map view filters
	BodiesVisible         bool
	SectorTypesVisible    bool
	ComponentNamesVisible bool

	// Game View state
	Renderer        *render.Renderer
	GameViewSurface *cairo.Surface
	GameViewBuffer  []uint8
}

func (e *Editor) State() *state.Edit {
	return &e.Edit
}

func NewEditor() *Editor {
	e := &Editor{
		Edit: state.Edit{
			MapView: state.MapView{
				Scale: 1.0,
				Step:  10,
				GridB: concepts.Vector2{1, 0},
			},
			Modified: false,
			DB:       concepts.NewEntityComponentDB(),
		},
		MapViewGrid:           MapViewGrid{Visible: true},
		BodiesVisible:         true,
		SectorTypesVisible:    false,
		ComponentNamesVisible: true,
	}
	e.Grid.IEditor = e
	e.EntityTree.IEditor = e
	e.MapViewGrid.Current = &e.Edit.MapView
	return e
}

func (e *Editor) ScreenToWorld(p *concepts.Vector2) *concepts.Vector2 {
	return p.Sub(e.Size.Mul(0.5)).MulSelf(1.0 / e.Scale).AddSelf(&e.Pos)
}

func (e *Editor) WorldToScreen(p *concepts.Vector2) *concepts.Vector2 {
	return p.Sub(&e.Pos).MulSelf(e.Scale).AddSelf(e.Size.Mul(0.5))
}

func (e *Editor) SetMapCursor(name string) {
	win, _ := e.MapArea.GetWindow()
	if name == "" {
		win.SetCursor(nil)
		return
	}
	dis, _ := gdk.DisplayGetDefault()
	cursor, _ := gdk.CursorNewFromName(dis, name)
	win.SetCursor(cursor)
}

func (e *Editor) UpdateTitle() {
	var title string
	if e.OpenFile != "" {
		title = e.OpenFile
	} else {
		title = "Untitled"
	}
	if e.Modified {
		title += " *"
	}
	e.Window.SetTitle(title)
}

func (e *Editor) UpdateStatus() {
	text := e.WorldGrid(&e.MouseWorld).StringHuman()
	if e.MousePressed {
		text = e.WorldGrid(&e.MouseDownWorld).StringHuman() + " -> " + text
		dist := e.WorldGrid(&e.MouseDownWorld).Sub(e.WorldGrid(&e.MouseWorld)).Length()
		text += " Length: " + strconv.FormatFloat(dist, 'f', 2, 64)
	}
	list := ""
	for _, obj := range e.HoveringObjects {
		if len(list) > 0 {
			list += ", "
		}
		switch v := obj.(type) {
		case *concepts.EntityRef:
			list += v.String()
		case *core.Segment:
			list += v.P.String()
		case *concepts.EntityComponentDB:
			list += "DB"
		}
	}
	text = list + " ( " + text + " )"
	e.StatusBar.SetText(text)
}

func (e *Editor) Integrate() {
	player := e.Renderer.Player()
	if player == nil {
		return
	}
	playerBody := core.BodyFromDb(player.Ref())

	if gameKeyMap[gdk.KEY_w] {
		controllers.MovePlayer(playerBody, playerBody.Angle)
	}
	if gameKeyMap[gdk.KEY_s] {
		controllers.MovePlayer(playerBody, playerBody.Angle+180.0)
	}
	if gameKeyMap[gdk.KEY_e] {
		controllers.MovePlayer(playerBody, playerBody.Angle+90.0)
	}
	if gameKeyMap[gdk.KEY_q] {
		controllers.MovePlayer(playerBody, playerBody.Angle+270.0)
	}
	if gameKeyMap[gdk.KEY_a] {
		playerBody.Angle -= constants.PlayerTurnSpeed * constants.TimeStepS
		playerBody.Angle = concepts.NormalizeAngle(playerBody.Angle)
	}
	if gameKeyMap[gdk.KEY_d] {
		playerBody.Angle += constants.PlayerTurnSpeed * constants.TimeStepS
		playerBody.Angle = concepts.NormalizeAngle(playerBody.Angle)
	}
	if gameKeyMap[gdk.KEY_space] {
		if playerBody.SectorEntityRef.Component(sectors.UnderwaterComponentIndex) != nil {
			playerBody.Force[2] += constants.PlayerSwimStrength
		} else if playerBody.OnGround {
			playerBody.Force[2] += constants.PlayerJumpForce
			playerBody.OnGround = false
		}
	}
	if gameKeyMap[gdk.KEY_c] {
		if playerBody.SectorEntityRef.Component(sectors.UnderwaterComponentIndex) != nil {
			playerBody.Force[2] -= constants.PlayerSwimStrength
		} else {
			player.Crouching = true
		}
	} else {
		player.Crouching = false
	}

	e.DB.NewControllerSet().ActGlobal(concepts.ControllerAlways)
	e.GatherHoveringObjects()
}

func (e *Editor) Load(filename string) {
	e.OpenFile = filename
	e.Modified = false
	e.UpdateTitle()
	db := concepts.NewEntityComponentDB()
	err := db.Load(e.OpenFile)
	if err != nil {
		e.Alert(fmt.Sprintf("Error loading world: %v", err))
		return
	}
	db.Simulation.Integrate = e.Integrate
	db.Simulation.Render = e.Window.QueueDraw
	e.DB = db
	e.SelectObjects([]any{}, true)
	e.GameView(e.GameArea.GetAllocatedWidth(), e.GameArea.GetAllocatedHeight())
	e.Grid.Refresh(e.SelectedObjects)
	e.EntityTree.Update()
}

func (e *Editor) Test() {
	e.Modified = false
	e.UpdateTitle()

	e.SelectObjects([]any{}, true)

	e.DB.Clear()
	controllers.CreateTestWorld2(e.DB)
	e.DB.Simulation.Integrate = e.Integrate
	e.DB.Simulation.Render = e.Window.QueueDraw

	e.GameView(e.GameArea.GetAllocatedWidth(), e.GameArea.GetAllocatedHeight())
	e.Grid.Refresh(e.SelectedObjects)
	e.EntityTree.Update()
}

func (e *Editor) ActionFinished(canceled bool) {
	e.UpdateTitle()
	controllers.AutoPortal(e.DB)
	if !canceled {
		e.UndoHistory = append(e.UndoHistory, e.CurrentAction)
		if len(e.UndoHistory) > 100 {
			e.UndoHistory = e.UndoHistory[(len(e.UndoHistory) - 100):]
		}
		e.RedoHistory = []state.IAction{}
	}
	e.Grid.Refresh(e.SelectedObjects)
	e.EntityTree.Update()
	e.SetMapCursor("")
	e.CurrentAction = nil
	e.ActTool()
}

func (e *Editor) NewAction(a state.IAction) {
	e.CurrentAction = a
}

func (e *Editor) ActTool() {
	switch e.Tool {
	case state.ToolSplitSegment:
		e.NewAction(&actions.SplitSegment{IEditor: e})
	case state.ToolSplitSector:
		e.NewAction(&actions.SplitSector{IEditor: e})
	case state.ToolAddSector:
		isector := archetypes.CreateSector(e.DB)
		s := core.SectorFromDb(isector)
		s.FloorMaterial = controllers.DefaultMaterial(e.DB)
		s.CeilMaterial = controllers.DefaultMaterial(e.DB)
		a := &actions.AddSector{Sector: s}
		a.AddBody.IEditor = e
		a.AddBody.EntityRef = isector
		e.NewAction(a)
	case state.ToolAddBody:
		body := archetypes.CreateBasic(e.DB, core.BodyComponentIndex)
		e.NewAction(&actions.AddBody{IEditor: e, EntityRef: body, Components: body.All()})
	case state.ToolAlignGrid:
		e.NewAction(&actions.AlignGrid{IEditor: e})
	default:
		return
	}
	e.CurrentAction.Act()
}

func (e *Editor) SwitchTool(tool state.EditorTool) {
	e.Tool = tool
	if e.CurrentAction != nil {
		e.CurrentAction.Cancel()
	} else {
		e.ActTool()
	}
}

func (e *Editor) UndoCurrent() {
	index := len(e.UndoHistory) - 1
	if index < 0 {
		return
	}
	a := e.UndoHistory[index]
	// Don't undo the current action!
	if a == e.CurrentAction {
		return
	}
	e.UndoHistory = e.UndoHistory[:index]
	if a == nil {
		return
	}
	a.Undo()
	controllers.AutoPortal(e.DB)
	e.Grid.Refresh(e.SelectedObjects)
	e.RedoHistory = append(e.RedoHistory, a)
}

func (e *Editor) RedoCurrent() {
	index := len(e.RedoHistory) - 1
	if index < 0 {
		return
	}
	a := e.RedoHistory[index]
	// Don't redo the current action!
	if a == e.CurrentAction {
		return
	}
	e.RedoHistory = e.RedoHistory[:index]
	if a == nil {
		return
	}
	a.Redo()
	controllers.AutoPortal(e.DB)
	e.Grid.Refresh(e.SelectedObjects)
	e.UndoHistory = append(e.UndoHistory, a)
}

func (e *Editor) SelectObjects(objects []any, updateTree bool) {
	e.SelectedObjects = objects
	e.Grid.Refresh(e.SelectedObjects)
	if updateTree {
		e.EntityTree.SetSelection(objects)
	}
}

func (e *Editor) Selecting() bool {
	_, ok := e.CurrentAction.(*actions.Select)
	return ok && e.MousePressed
}

func (e *Editor) SelectionBox() (v1 *concepts.Vector2, v2 *concepts.Vector2) {
	// Copy
	mw := e.MouseWorld
	mdw := e.MouseDownWorld
	v1 = &mw
	v2 = &mdw

	if e.MousePressed && v2[0] < v1[0] {
		v1[0], v2[0] = v2[0], v1[0]
	}
	if e.MousePressed && v2[1] < v1[1] {
		v1[1], v2[1] = v2[1], v1[1]
	}
	return
}

func (e *Editor) GatherHoveringObjects() {
	// Hovering
	v1, v2 := e.SelectionBox()

	e.HoveringObjects = []any{}

	for _, isector := range e.DB.All(core.SectorComponentIndex) {
		sector := isector.(*core.Sector)

		for _, segment := range sector.Segments {
			if e.CurrentAction == nil {
				if e.Mouse.Sub(e.WorldToScreen(&segment.P)).Length() < state.SegmentSelectionEpsilon {
					e.HoveringObjects = append(e.HoveringObjects, segment)
				}
			} else if editor.Selecting() {
				if segment.P[0] >= v1[0] && segment.P[1] >= v1[1] && segment.P[0] <= v2[0] && segment.P[1] <= v2[1] {
					if state.IndexOf(e.HoveringObjects, segment) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, segment)
					}
				}
				/*if segment.AABBIntersect(v1[0], v1[1], v2[0], v2[1]) {
					if state.IndexOf(e.HoveringObjects, segment) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, segment)
					}
				}*/
			}
		}

		if e.Selecting() {
			for _, ibody := range sector.Bodies {
				body := core.BodyFromDb(ibody)
				p := body.Pos.Original
				if p[0]+body.BoundingRadius >= v1[0] && p[0]-body.BoundingRadius <= v2[0] &&
					p[1]+body.BoundingRadius >= v1[1] && p[1]-body.BoundingRadius <= v2[1] {
					if state.IndexOf(e.HoveringObjects, ibody) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, ibody)
					}
				}
			}
		}
	}
}

func (e *Editor) GameView(w, h int) {
	e.Renderer = render.NewRenderer()
	e.Renderer.DB = e.DB
	e.Renderer.ScreenWidth = w
	e.Renderer.ScreenHeight = h
	e.Renderer.Initialize()
	_, _ = render.NewFont("/Library/Fonts/Courier New.ttf", 24)
	e.GameViewSurface = cairo.CreateImageSurface(cairo.FORMAT_ARGB32, w, h)
	// We'll need the raw buffer to draw into, but we'll use
	// a bit of Go magic to get some type safety back...
	length := w * h * 4
	e.GameViewSurface.Flush() // necessary?
	pBuffer := e.GameViewSurface.GetData()
	// Make a new slice
	e.GameViewBuffer = unsafe.Slice((*uint8)(pBuffer), length)
}

func (e *Editor) AddSimpleMenuAction(name string, cb func(obj *glib.Object)) {
	action := glib.SimpleActionNew(name, nil)
	action.Connect("activate", func(sa *glib.SimpleAction) { cb(sa.Object) })
	e.App.AddAction(action)
}

func (e *Editor) MoveSurface(delta float64, floor bool, slope bool) {
	action := &actions.MoveSurface{IEditor: e, Delta: delta, Floor: floor, Slope: slope}
	e.NewAction(action)
	action.Act()
}

func (e *Editor) Alert(text string) {
	if win, err := e.Container.GetToplevel(); err == nil {
		dlg := gtk.MessageDialogNew(win.(gtk.IWindow), gtk.DIALOG_DESTROY_WITH_PARENT, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, text)
		dlg.Connect("response", func() {
			dlg.Destroy()
		})
		dlg.ShowAll()
	}
}
