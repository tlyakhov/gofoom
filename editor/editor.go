package main

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"

	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/sectors"

	"github.com/gotk3/gotk3/gdk"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/properties"
	"tlyakhov/gofoom/editor/state"
	"tlyakhov/gofoom/render"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/controllers"
)

type EditorWidgets struct {
	App         *gtk.Application
	Window      *gtk.ApplicationWindow
	GameArea    *gtk.DrawingArea
	MapArea     *gtk.DrawingArea
	MobTypes    *gtk.ComboBoxText
	SectorTypes *gtk.ComboBoxText
	StatusBar   *gtk.Label
}

type Editor struct {
	state.Edit
	// What we're editing.

	MapViewGrid
	EditorWidgets
	properties.Grid

	// Map view filters
	MobsVisible        bool
	SectorTypesVisible bool
	MobTypesVisible    bool

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
		MapViewGrid:        MapViewGrid{Visible: true},
		MobsVisible:        true,
		SectorTypesVisible: false,
		MobTypesVisible:    true,
	}
	e.DB.Simulation.Integrate = e.Integrate
	e.DB.Simulation.Render = e.Window.QueueDraw
	e.Grid.IEditor = e
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
		list += obj.Entity
	}
	text = list + " ( " + text + " )"
	e.StatusBar.SetText(text)
}

func (e *Editor) Integrate() {
	player := e.Renderer.Player()
	playerMob := core.MobFromDb(player.EntityRef())

	if gameKeyMap[gdk.KEY_w] {
		controllers.MovePlayer(playerMob, playerMob.Angle)
	}
	if gameKeyMap[gdk.KEY_s] {
		controllers.MovePlayer(playerMob, playerMob.Angle+180.0)
	}
	if gameKeyMap[gdk.KEY_e] {
		controllers.MovePlayer(playerMob, playerMob.Angle+90.0)
	}
	if gameKeyMap[gdk.KEY_q] {
		controllers.MovePlayer(playerMob, playerMob.Angle+270.0)
	}
	if gameKeyMap[gdk.KEY_a] {
		playerMob.Angle -= constants.PlayerTurnSpeed * constants.TimeStepS
		playerMob.Angle = concepts.NormalizeAngle(playerMob.Angle)
	}
	if gameKeyMap[gdk.KEY_d] {
		playerMob.Angle += constants.PlayerTurnSpeed * constants.TimeStepS
		playerMob.Angle = concepts.NormalizeAngle(playerMob.Angle)
	}
	if gameKeyMap[gdk.KEY_space] {
		if playerMob.SectorEntityRef.Component(sectors.UnderwaterComponentIndex) != nil {
			playerMob.Force[2] += constants.PlayerSwimStrength
		} else if playerMob.OnGround {
			playerMob.Force[2] += constants.PlayerJumpForce
			playerMob.OnGround = false
		}
	}
	if gameKeyMap[gdk.KEY_c] {
		if playerMob.SectorEntityRef.Component(sectors.UnderwaterComponentIndex) != nil {
			playerMob.Force[2] -= constants.PlayerSwimStrength
		} else {
			player.Crouching = true
		}
	} else {
		player.Crouching = false
	}

	e.DB.NewControllerSet().ActGlobal("Always")
	e.GatherHoveringObjects()
}

func (e *Editor) Load(filename string) {
	e.OpenFile = filename
	e.Modified = false
	e.UpdateTitle()
	e.DB.Clear()
	err := e.DB.Load(e.OpenFile)
	if err != nil {
		e.Alert(fmt.Sprintf("Error loading world: %v", err))
		return
	}

	e.SelectObjects([]any{})
	e.GameView(e.GameArea.GetAllocatedWidth(), e.GameArea.GetAllocatedHeight())
	e.Grid.Refresh(e.SelectedObjects)
}

func (e *Editor) Test() {
	e.Modified = false
	e.UpdateTitle()

	e.SelectObjects([]any{})

	e.DB.Clear()
	controllers.CreateTestWorld(e.DB)
	e.GameView(e.GameArea.GetAllocatedWidth(), e.GameArea.GetAllocatedHeight())
	e.Grid.Refresh(e.SelectedObjects)

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
		typeId := e.SectorTypes.GetActiveID()
		t := concepts.DB().AllTypes[typeId]
		s := reflect.New(t).Interface().(core.Sector)
		s.Construct(nil)
		if simmed, ok := s.(core.Simulated); ok {
			simmed.Attach(e.World.Sim())
		}
		s.FloorMaterial = e.World.DefaultMaterial()
		s.CeilMaterial = e.World.DefaultMaterial()
		s.SetParent(e.World.Map)
		e.NewAction(&actions.AddSector{IEditor: e, Sector: s})
	case state.ToolAddMob:
		typeId := e.MobTypes.GetActiveID()
		t := concepts.DB().AllTypes[typeId]
		ae := reflect.New(t).Interface().(core.Mob)
		ae.Construct(nil)
		if simmed, ok := ae.(core.Simulated); ok {
			simmed.Attach(e.World.Sim())
		}
		e.NewAction(&actions.AddEntity{IEditor: e, Mob: ae})
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

func (e *Editor) Undo() {
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

func (e *Editor) Redo() {
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

func (e *Editor) SelectObjects(objects []any) {
	if len(objects) == 0 {
		objects = append(objects, e.World)
	}

	e.SelectedObjects = objects
	e.Grid.Refresh(e.SelectedObjects)
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
					mp := &state.MapPoint{Segment: segment}
					if concepts.IndexOf(e.HoveringObjects, mp) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, mp)
					}
				}
				if segment.AABBIntersect(v1[0], v1[1], v2[0], v2[1]) {
					if concepts.IndexOf(e.HoveringObjects, segment) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, segment)
					}
				}
			}
		}

		if e.Selecting() {
			for _, mob := range sector.Mobs {
				pe := mob
				p := pe.Pos.Original
				if p[0]+pe.BoundingRadius >= v1[0] && p[0]-pe.BoundingRadius <= v2[0] &&
					p[1]+pe.BoundingRadius >= v1[1] && p[1]-pe.BoundingRadius <= v2[1] {
					if concepts.IndexOf(e.HoveringObjects, mob) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, mob)
					}
				}
			}
		}
	}
}

func (e *Editor) GameView(w, h int) {
	e.Renderer = render.NewRenderer()
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
