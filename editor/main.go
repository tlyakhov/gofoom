package main

import (
	"fmt"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	_ "github.com/tlyakhov/gofoom/behaviors"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/logic"
	"github.com/tlyakhov/gofoom/logic/entity"
	"github.com/tlyakhov/gofoom/render"

	_ "github.com/tlyakhov/gofoom/logic/provide"
	_ "github.com/tlyakhov/gofoom/logic/sector"
)

const (
	KEY_LEFT  uint = 65361
	KEY_UP    uint = 65362
	KEY_RIGHT uint = 65363
	KEY_DOWN  uint = 65364

	GridSize                float64 = 10
	SegmentSelectionEpsilon float64 = 5.0
)

var (
	ColorSelectionPrimary   concepts.Vector3 = concepts.Vector3{0, 1, 0}
	ColorSelectionSecondary concepts.Vector3 = concepts.Vector3{0, 1, 1}
	ColorPVS                concepts.Vector3 = concepts.Vector3{0.6, 1, 0.6}
	Scale                   float64          = 1.0
	Pos                     concepts.Vector2
	Mouse                   concepts.Vector2
	GameMap                 *logic.MapService
	GridVisible             bool = true
	EntitiesVisible         bool = true
	MapViewSize             concepts.Vector2
	HoveringObjects         map[string]concepts.ISerializable
	SelectedObjects         map[string]concepts.ISerializable
)

func ScreenToWorld(p concepts.Vector2) concepts.Vector2 {
	return p.Sub(MapViewSize.Mul(0.5)).Mul(1.0 / Scale).Add(Pos)
}

func WorldToScreen(p concepts.Vector2) concepts.Vector2 {
	return p.Sub(Pos).Mul(Scale).Add(MapViewSize.Mul(0.5))
}

func DrawSector(cr *cairo.Context, sector core.AbstractSector) {
	phys := sector.Physical()

	if len(phys.Segments) == 0 {
		return
	}

	sectorHovering := HoveringObjects[sector.GetBase().ID] == sector
	sectorSelected := SelectedObjects[sector.GetBase().ID] == sector

	if EntitiesVisible {
		// Draw entities
	}

	for i, segment := range phys.Segments {
		next := phys.Segments[(i+1)%len(phys.Segments)]

		segmentHovering := HoveringObjects[segment.GetBase().ID] == segment
		segmentSelected := SelectedObjects[segment.GetBase().ID] == segment

		if segment.AdjacentSector == nil {
			cr.SetSourceRGB(1, 1, 1)
		} else {
			cr.SetSourceRGB(1, 1, 0)
		}

		if sectorHovering || sectorSelected {
			if segment.AdjacentSector == nil {
				cr.SetSourceRGB(ColorSelectionPrimary.X, ColorSelectionPrimary.Y, ColorSelectionPrimary.Z)
			} else {
				cr.SetSourceRGB(ColorSelectionSecondary.X, ColorSelectionSecondary.Y, ColorSelectionSecondary.Z)
			}
		} else if segmentHovering {
			cr.SetSourceRGB(ColorSelectionSecondary.X, ColorSelectionSecondary.Y, ColorSelectionSecondary.Z)
		} else if segmentSelected {
			cr.SetSourceRGB(ColorSelectionPrimary.X, ColorSelectionPrimary.Y, ColorSelectionPrimary.Z)
		}

		// Highlight PVS sectors...
		for _, obj := range SelectedObjects {
			s2, ok := obj.(core.AbstractSector)
			if !ok || s2 == sector {
				continue
			}
			if s2.Physical().PVS[sector.GetBase().ID] != nil {
				cr.SetSourceRGB(ColorPVS.X, ColorPVS.Y, ColorPVS.Z)
			}
		}

		// Draw segment
		cr.NewPath()
		cr.MoveTo(segment.A.X, segment.A.Y)
		cr.LineTo(next.A.X, next.A.Y)
		cr.ClosePath()
		cr.Stroke()
		// Draw normal
		cr.NewPath()
		ns := next.A.Add(segment.A).Mul(0.5)
		ne := ns.Add(segment.Normal.Mul(4.0))
		cr.MoveTo(ns.X, ns.Y)
		cr.LineTo(ne.X, ne.Y)
		cr.ClosePath()
		cr.Stroke()
	}
}

func main() {
	gtk.Init(nil)

	// gui boilerplate
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	map_area, _ := gtk.DrawingAreaNew()
	map_area.SetEvents(int(gdk.POINTER_MOTION_MASK))
	game_area, _ := gtk.DrawingAreaNew()
	prop_grid, _ := gtk.GridNew()
	hpane, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	vpane, _ := gtk.PanedNew(gtk.ORIENTATION_VERTICAL)
	hpane.Pack1(map_area, true, true)
	hpane.Pack2(vpane, true, true)
	vpane.Pack1(game_area, true, true)
	vpane.Pack2(prop_grid, true, true)
	win.Add(hpane)
	win.SetSizeRequest(800, 600)
	win.SetTitle("Arrow keys")
	win.Connect("destroy", gtk.MainQuit)
	win.ShowAll()

	renderer := render.NewRenderer()
	renderer.ScreenWidth = 800
	renderer.ScreenHeight = 600
	renderer.Initialize()
	GameMap = logic.LoadMap("data/classicMap.json")
	ps := entity.NewPlayerService(GameMap.Player.(*entities.Player))
	ps.Collide()
	renderer.Map = GameMap.Map
	_, _ = render.NewFont("/Library/Fonts/Courier New.ttf", 24)

	// Data
	x := 0.0
	y := 0.0
	keyMap := map[uint]func(){
		KEY_LEFT:  func() { x-- },
		KEY_UP:    func() { y-- },
		KEY_RIGHT: func() { x++ },
		KEY_DOWN:  func() { y++ },
	}

	// Event handlers
	map_area.Connect("draw", func(da *gtk.DrawingArea, cr *cairo.Context) {
		//cr.SetSourceRGB(0, 0, 0)
		//cr.Rectangle(x*unitSize, y*unitSize, unitSize, unitSize)
		//cr.Fill()
		w := da.GetAllocatedWidth()
		h := da.GetAllocatedHeight()
		MapViewSize = concepts.Vector2{float64(w), float64(h)}

		//cr.SetMatrix(cairo.NewMatrix(1, 0, 0, 1, 0, 0))
		cr.SetSourceRGB(0, 0, 0)
		cr.Paint()
		t := Pos.Mul(-Scale).Add(MapViewSize.Mul(0.5))
		cr.Translate(t.X, t.Y)
		cr.Scale(Scale, Scale)

		if GridVisible && Scale*GridSize > 5.0 {
			start := ScreenToWorld(concepts.Vector2{}).Mul(1.0 / GridSize).Floor().Mul(GridSize)
			end := ScreenToWorld(MapViewSize).Mul(1.0 / GridSize).Add(concepts.Vector2{1, 1}).Floor().Mul(GridSize)

			cr.SetSourceRGB(0.5, 0.5, 0.5)
			for x := start.X; x < end.Y; x += GridSize {
				for y := start.Y; y < end.Y; y += GridSize {
					cr.Rectangle(x, y, 1, 1)
					cr.Fill()
				}
			}
		}

		for _, sector := range GameMap.Sectors {
			DrawSector(cr, sector)
		}
		cr.ShowText(fmt.Sprintf("%v, %v", Mouse.X, Mouse.Y))
	})

	map_area.Connect("motion-notify-event", func(da *gtk.DrawingArea, ev *gdk.Event) {
		motion := &gdk.EventMotion{ev}
		Mouse.X, Mouse.Y = motion.MotionVal()
		win.QueueDraw()
	})
	win.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{ev}
		if move, found := keyMap[keyEvent.KeyVal()]; found {
			move()
			win.QueueDraw()
		}
	})

	gtk.Main()
}
