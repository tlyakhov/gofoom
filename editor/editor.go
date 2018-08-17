package main

import (
	"reflect"
	"unsafe"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/tlyakhov/gofoom/render"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/logic"
)

type MapViewState struct {
	Scale       float64
	Pos         concepts.Vector2 // World
	MapViewSize concepts.Vector2 // Screen
}

type MapViewGrid struct {
	Prev    MapViewState
	Visible bool
	Surface *cairo.Surface
}

type EditorWidgets struct {
	App          *gtk.Application
	Window       *gtk.ApplicationWindow
	PropertyGrid *gtk.Grid
	GameArea     *gtk.DrawingArea
	MapArea      *gtk.DrawingArea
}

type Editor struct {
	// What we're editing.
	GameMap *logic.MapService
	MapViewState
	Grid MapViewGrid
	EditorWidgets

	// Map view positions in world/screen space.
	Mouse          concepts.Vector2 // Screen
	MouseDown      concepts.Vector2 // Screen
	MouseWorld     concepts.Vector2
	MouseDownWorld concepts.Vector2
	MousePressed   bool

	// Map view filters
	EntitiesVisible    bool
	SectorTypesVisible bool
	HoveringObjects    []concepts.ISerializable
	SelectedObjects    []concepts.ISerializable

	// Should be typed enum, used by actions only?
	State         string
	CurrentAction AbstractAction
	UndoHistory   []AbstractAction
	RedoHistory   []AbstractAction

	// Game View state
	Renderer        *render.Renderer
	GameViewSurface *cairo.Surface
	GameViewBuffer  []uint8
}

func NewEditor() *Editor {
	return &Editor{
		MapViewState:       MapViewState{Scale: 1.0},
		Grid:               MapViewGrid{Visible: true},
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

func (e *Editor) SelectObjects(objects []concepts.ISerializable) {
	if len(objects) == 0 {
		objects = append(objects, e.GameMap)
	}

	e.SelectedObjects = objects
	e.RefreshPropertyGrid()
}

func indexOfObject(s []concepts.ISerializable, obj concepts.ISerializable) int {
	id := obj.GetBase().ID
	for i, e := range s {
		if e.GetBase().ID == id && reflect.TypeOf(obj) == reflect.TypeOf(e) {
			return i
		}
	}
	return -1
}

func (e *Editor) Selecting() bool {
	_, ok := e.CurrentAction.(*SelectAction)
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

	for _, sector := range e.GameMap.Sectors {
		phys := sector.Physical()

		for _, segment := range phys.Segments {
			if e.CurrentAction == nil {
				if e.Mouse.Sub(e.WorldToScreen(segment.A)).Length() < SegmentSelectionEpsilon {
					e.HoveringObjects = append(e.HoveringObjects, segment)
				}
			} else if editor.Selecting() {
				if segment.A.X >= v1.X && segment.A.Y >= v1.Y && segment.A.X <= v2.X && segment.A.Y <= v2.Y {
					mp := &MapPoint{Segment: segment}
					if indexOfObject(e.HoveringObjects, mp) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, mp)
					}
				}
				if segment.AABBIntersect(v1.X, v1.Y, v2.X, v2.Y) {
					if indexOfObject(e.HoveringObjects, segment) == -1 {
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
					if indexOfObject(e.HoveringObjects, entity) == -1 {
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
	e.Renderer.Map = e.GameMap.Map
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
