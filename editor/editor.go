package main

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/tlyakhov/gofoom/editor/actions"

	"github.com/gotk3/gotk3/gdk"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/editor/properties"
	"github.com/tlyakhov/gofoom/editor/state"
	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/logic/entity"
	"github.com/tlyakhov/gofoom/registry"
	"github.com/tlyakhov/gofoom/render"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/logic"
)

type EditorWidgets struct {
	App         *gtk.Application
	Window      *gtk.ApplicationWindow
	GameArea    *gtk.DrawingArea
	MapArea     *gtk.DrawingArea
	EntityTypes *gtk.ComboBoxText
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

func (e *Editor) ScreenToWorld(p concepts.Vector2) concepts.Vector2 {
	return p.Sub(e.Size.Mul(0.5)).Mul(1.0 / e.Scale).Add(e.Pos)
}

func (e *Editor) WorldToScreen(p concepts.Vector2) concepts.Vector2 {
	return p.Sub(e.Pos).Mul(e.Scale).Add(e.Size.Mul(0.5))
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
	text := e.WorldGrid(e.MouseWorld).StringHuman()
	if e.MousePressed {
		text = e.WorldGrid(e.MouseDownWorld).StringHuman() + " -> " + text
		dist := e.WorldGrid(e.MouseDownWorld).Sub(e.WorldGrid(e.MouseWorld)).Length()
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

func (e *Editor) Load(filename string) {
	e.OpenFile = filename
	e.Modified = false
	e.UpdateTitle()
	e.World = logic.LoadMap(e.OpenFile)
	ps := entity.NewPlayerService(e.World.Player.(*entities.Player))
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
		s := &core.PhysicalSector{}
		s.Initialize()
		s.FloorMaterial = e.World.DefaultMaterial()
		s.CeilMaterial = e.World.DefaultMaterial()
		s.SetParent(e.World.Map)
		e.NewAction(&actions.AddSector{IEditor: e, Sector: s})
	case state.ToolAddEntity:
		typeId := e.EntityTypes.GetActiveID()
		t := registry.Instance().All[typeId]
		fmt.Println(t)
		ae := reflect.New(t).Interface().(core.AbstractEntity)
		ae.Initialize()
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

func (e *Editor) SelectionBox() (v1 concepts.Vector2, v2 concepts.Vector2) {
	v1 = e.MouseWorld
	v2 = e.MouseDownWorld

	if e.MousePressed && v2.X < v1.X {
		tmp := v1.X
		v1.X = v2.X
		v2.X = tmp
	}
	if e.MousePressed && v2.Y < v1.Y {
		tmp := v1.Y
		v1.Y = v2.Y
		v2.Y = tmp
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
				if e.Mouse.Sub(e.WorldToScreen(segment.P)).Length() < state.SegmentSelectionEpsilon {
					e.HoveringObjects = append(e.HoveringObjects, segment)
				}
			} else if editor.Selecting() {
				if segment.P.X >= v1.X && segment.P.Y >= v1.Y && segment.P.X <= v2.X && segment.P.Y <= v2.Y {
					mp := &state.MapPoint{Segment: segment}
					if concepts.IndexOf(e.HoveringObjects, mp) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, mp)
					}
				}
				if segment.AABBIntersect(v1.X, v1.Y, v2.X, v2.Y) {
					if concepts.IndexOf(e.HoveringObjects, segment) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, segment)
					}
				}
			}
		}

		if e.Selecting() {
			for _, entity := range sector.Physical().Entities {
				pe := entity.Physical()
				if pe.Pos.X+pe.BoundingRadius >= v1.X && pe.Pos.X-pe.BoundingRadius <= v2.X &&
					pe.Pos.Y+pe.BoundingRadius >= v1.Y && pe.Pos.Y-pe.BoundingRadius <= v2.Y {
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
	header := reflect.SliceHeader{uintptr(pBuffer), length, length}
	// Type safe!
	e.GameViewBuffer = *(*[]uint8)(unsafe.Pointer(&header))
}

func (e *Editor) AddSimpleMenuAction(name string, cb func(obj *glib.Object)) {
	action := glib.SimpleActionNew(name, nil)
	action.Connect("activate", cb)
	e.App.AddAction(action)
}

func (e *Editor) MoveSurface(delta float64, floor bool, slope bool) {
	action := &actions.MoveSurface{IEditor: e, Delta: delta, Floor: floor, Slope: slope}
	e.NewAction(action)
	action.Act()
}
