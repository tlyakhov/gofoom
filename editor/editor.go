package main

import (
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

	"tlyakhov/gofoom/controllers/entity"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/editor/properties"
	"tlyakhov/gofoom/editor/state"
	"tlyakhov/gofoom/entities"
	"tlyakhov/gofoom/registry"
	"tlyakhov/gofoom/render"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/controllers"
)

type EditorWidgets struct {
	App         *gtk.Application
	Window      *gtk.ApplicationWindow
	GameArea    *gtk.DrawingArea
	MapArea     *gtk.DrawingArea
	EntityTypes *gtk.ComboBoxText
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
	EntitiesVisible    bool
	SectorTypesVisible bool
	EntityTypesVisible bool

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
		},
		MapViewGrid:        MapViewGrid{Visible: true},
		EntitiesVisible:    true,
		SectorTypesVisible: false,
		EntityTypesVisible: true,
	}
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
		list += obj.GetBase().ID
	}
	text = list + " ( " + text + " )"
	e.StatusBar.SetText(text)
}

func (e *Editor) Integrate() {
	ps := entity.NewPlayerController(editor.World.Player.(*entities.Player))

	if gameKeyMap[gdk.KEY_w] {
		ps.Move(ps.Player.Angle)
	}
	if gameKeyMap[gdk.KEY_s] {
		ps.Move(ps.Player.Angle + 180.0)
	}
	if gameKeyMap[gdk.KEY_e] {
		ps.Move(ps.Player.Angle + 90.0)
	}
	if gameKeyMap[gdk.KEY_q] {
		ps.Move(ps.Player.Angle + 270.0)
	}
	if gameKeyMap[gdk.KEY_a] {
		ps.Player.Angle -= constants.PlayerTurnSpeed * constants.TimeStep
		ps.Player.Angle = concepts.NormalizeAngle(ps.Player.Angle)
	}
	if gameKeyMap[gdk.KEY_d] {
		ps.Player.Angle += constants.PlayerTurnSpeed * constants.TimeStep
		ps.Player.Angle = concepts.NormalizeAngle(ps.Player.Angle)
	}
	if gameKeyMap[gdk.KEY_space] {
		if _, ok := ps.Player.Sector.(*sectors.Underwater); ok {
			ps.Player.Vel.Now[2] += constants.PlayerSwimStrength * constants.TimeStep
		} else if ps.Player.OnGround {
			ps.Player.Vel.Now[2] += constants.PlayerJumpStrength * constants.TimeStep
			ps.Player.OnGround = false
		}
	}
	if gameKeyMap[gdk.KEY_c] {
		if _, ok := ps.Player.Sector.(*sectors.Underwater); ok {
			ps.Player.Vel.Now[2] -= constants.PlayerSwimStrength * constants.TimeStep
		} else {
			ps.Crouching = true
		}
	} else {
		ps.Crouching = false
	}

	editor.World.Frame()
	editor.GatherHoveringObjects()
}

func (e *Editor) Load(filename string) {
	e.OpenFile = filename
	e.Modified = false
	e.UpdateTitle()
	sim := core.NewSimulation()
	sim.Integrate = e.Integrate
	sim.Render = e.Window.QueueDraw
	e.World = controllers.LoadMap(e.OpenFile)
	e.World.Attach(sim)
	ps := entity.NewPlayerController(e.World.Player.(*entities.Player))
	ps.Collide()
	e.SelectObjects([]concepts.ISerializable{})
	e.GameView(e.GameArea.GetAllocatedWidth(), e.GameArea.GetAllocatedHeight())
	e.Grid.Refresh(e.SelectedObjects)
}

func (e *Editor) Test() {
	e.Modified = false
	e.UpdateTitle()
	sim := core.NewSimulation()
	sim.Integrate = e.Integrate
	sim.Render = e.Window.QueueDraw
	e.World = controllers.NewMapController(new(core.Map))
	e.World.Initialize()
	e.World.Attach(sim)
	e.World.CreateTest()
	ps := entity.NewPlayerController(e.World.Player.(*entities.Player))
	ps.Collide()
	e.SelectObjects([]concepts.ISerializable{})
	e.GameView(e.GameArea.GetAllocatedWidth(), e.GameArea.GetAllocatedHeight())
	e.Grid.Refresh(e.SelectedObjects)

}

func (e *Editor) ActionFinished(canceled bool) {
	e.UpdateTitle()
	e.World.AutoPortal()
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
		t := registry.Instance().All[typeId]
		s := reflect.New(t).Interface().(core.AbstractSector)
		s.Initialize()
		if simmed, ok := s.(core.Simulated); ok {
			simmed.Attach(e.World.Sim())
		}
		s.Physical().FloorMaterial = e.World.DefaultMaterial()
		s.Physical().CeilMaterial = e.World.DefaultMaterial()
		s.SetParent(e.World.Map)
		e.NewAction(&actions.AddSector{IEditor: e, Sector: s})
	case state.ToolAddEntity:
		typeId := e.EntityTypes.GetActiveID()
		t := registry.Instance().All[typeId]
		ae := reflect.New(t).Interface().(core.AbstractEntity)
		ae.Initialize()
		if simmed, ok := ae.(core.Simulated); ok {
			simmed.Attach(e.World.Sim())
		}
		e.NewAction(&actions.AddEntity{IEditor: e, Entity: ae})
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
	e.World.AutoPortal()
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
	e.World.AutoPortal()
	e.Grid.Refresh(e.SelectedObjects)
	e.UndoHistory = append(e.UndoHistory, a)
}

func (e *Editor) SelectObjects(objects []concepts.ISerializable) {
	if len(objects) == 0 {
		objects = append(objects, e.World.Map)
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

	e.HoveringObjects = []concepts.ISerializable{}

	for _, sector := range e.World.Sectors {
		phys := sector.Physical()

		for _, segment := range phys.Segments {
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
			for _, entity := range sector.Physical().Entities {
				pe := entity.Physical()
				p := pe.Pos.Original
				if p[0]+pe.BoundingRadius >= v1[0] && p[0]-pe.BoundingRadius <= v2[0] &&
					p[1]+pe.BoundingRadius >= v1[1] && p[1]-pe.BoundingRadius <= v2[1] {
					if concepts.IndexOf(e.HoveringObjects, entity) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, entity)
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
	e.Renderer.Map = e.World.Map
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
