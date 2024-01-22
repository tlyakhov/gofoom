package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/editor/actions"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/sectors"
	"tlyakhov/gofoom/editor/properties"
	"tlyakhov/gofoom/editor/state"
	"tlyakhov/gofoom/render"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/controllers"
)

type EditorWidgets struct {
	App          fyne.App
	Window       fyne.Window
	LabelStatus  *widget.Label
	PropertyGrid *fyne.Container
	GameWidget   *GameWidget
	MapWidget    *MapWidget
}

type Editor struct {
	state.Edit
	// What we're editing.

	MapViewGrid
	EditorWidgets
	EditorMenu
	EntityList
	properties.Grid

	// Game View state
	Renderer       *render.Renderer
	MapViewSurface *image.RGBA
}

func (e *Editor) State() *state.Edit {
	return &e.Edit
}

func NewEditor() *Editor {
	e := &Editor{
		Edit: state.Edit{
			DB: concepts.NewEntityComponentDB(),
			MapView: state.MapView{
				Scale: 1.0,
				Step:  10,
				GridB: concepts.Vector2{1, 0},
			},
			Modified:              false,
			BodiesVisible:         true,
			SectorTypesVisible:    false,
			ComponentNamesVisible: true,
		},
		MapViewGrid: MapViewGrid{Visible: true},
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

func (e *Editor) SetMapCursor(cursor desktop.Cursor) {
	e.MapWidget.MapCursor = cursor
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
	e.LabelStatus.SetText(text)
}

func (e *Editor) Integrate() {
	player := e.Renderer.Player
	if player == nil {
		return
	}
	playerBody := core.BodyFromDb(player.Ref())

	if e.GameWidget.KeyMap["W"] {
		controllers.MovePlayer(playerBody, playerBody.Angle.Now, e.DB.EditorPaused)
	}
	if e.GameWidget.KeyMap["S"] {
		controllers.MovePlayer(playerBody, playerBody.Angle.Now+180.0, e.DB.EditorPaused)
	}
	if e.GameWidget.KeyMap["E"] {
		controllers.MovePlayer(playerBody, playerBody.Angle.Now+90.0, e.DB.EditorPaused)
	}
	if e.GameWidget.KeyMap["Q"] {
		controllers.MovePlayer(playerBody, playerBody.Angle.Now+270.0, e.DB.EditorPaused)
	}
	if e.GameWidget.KeyMap["A"] {
		playerBody.Angle.Now -= constants.PlayerTurnSpeed * constants.TimeStepS
		playerBody.Angle.Now = concepts.NormalizeAngle(playerBody.Angle.Now)
	}
	if e.GameWidget.KeyMap["D"] {
		playerBody.Angle.Now += constants.PlayerTurnSpeed * constants.TimeStepS
		playerBody.Angle.Now = concepts.NormalizeAngle(playerBody.Angle.Now)
	}
	if e.GameWidget.KeyMap["Space"] {
		if playerBody.SectorEntityRef.Component(sectors.UnderwaterComponentIndex) != nil {
			playerBody.Force[2] += constants.PlayerSwimStrength
		} else if playerBody.OnGround {
			playerBody.Force[2] += constants.PlayerJumpForce
			playerBody.OnGround = false
		}
	}
	if e.GameWidget.KeyMap["C"] {
		if playerBody.SectorEntityRef.Component(sectors.UnderwaterComponentIndex) != nil {
			playerBody.Force[2] -= constants.PlayerSwimStrength
		} else {
			player.Crouching = true
		}
	} else {
		player.Crouching = false
	}

	e.DB.ActAllControllers(concepts.ControllerAlways)
	e.GatherHoveringObjects()
}

func (e *Editor) ChangeSelectedTransformables(m *concepts.Matrix2) {
	for _, t := range e.State().SelectedTransformables {
		switch target := t.(type) {
		case *concepts.Vector2:
			m.ProjectSelf(target)
		case *concepts.Vector3:
			m.ProjectXZSelf(target)
		case *concepts.Matrix2:
			*target = *m.Mul(target)
			//			log.Printf("CST: %v", target.StringHuman())
		}
	}
}

func (e *Editor) Load(filename string) {
	e.Lock.Lock()
	defer e.Lock.Unlock()
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
	db.Simulation.Render = e.GameWidget.Draw
	e.DB = db
	e.SelectObjects([]any{}, true)
	editor.ResizeRenderer(640, 360)
	e.Grid.Refresh(e.SelectedObjects)
	e.EntityList.Update()
}

func (e *Editor) Test() {
	e.Lock.Lock()
	defer e.Lock.Unlock()

	e.Modified = false
	e.UpdateTitle()

	e.SelectObjects([]any{}, true)

	e.DB.Clear()
	controllers.CreateTestWorld2(e.DB)
	e.DB.Simulation.Integrate = e.Integrate
	e.DB.Simulation.Render = e.GameWidget.Draw

	e.ResizeRenderer(640, 360)
	e.Grid.Refresh(e.SelectedObjects)
	e.EntityList.Update()
}

func (e *Editor) ActionFinished(canceled, refreshProperties, autoPortal bool) {
	e.UpdateTitle()
	if autoPortal {
		e.Lock.Lock()
		controllers.AutoPortal(e.DB)
		e.Lock.Unlock()
	}
	if !canceled {
		e.UndoHistory = append(e.UndoHistory, e.CurrentAction)
		if len(e.UndoHistory) > 100 {
			e.UndoHistory = e.UndoHistory[(len(e.UndoHistory) - 100):]
		}
		e.RedoHistory = []state.IAction{}
	}
	if refreshProperties {
		e.Grid.Refresh(e.SelectedObjects)
		e.EntityList.Update()
	}
	e.SetMapCursor(desktop.DefaultCursor)
	e.CurrentAction = nil
	go e.ActTool()
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
		s.FloorSurface.Material = controllers.DefaultMaterial(e.DB)
		s.CeilSurface.Material = controllers.DefaultMaterial(e.DB)
		a := &actions.AddSector{Sector: s}
		a.AddBody.IEditor = e
		a.AddBody.EntityRef = isector
		a.AddBody.Components = isector.All()
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
	e.Lock.Lock()
	defer e.Lock.Unlock()
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
	e.Lock.Lock()
	defer e.Lock.Unlock()

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

func (e *Editor) SelectObjects(objects []any, updateEntityList bool) {
	e.SelectedObjects = objects
	e.Grid.Refresh(e.SelectedObjects)
	if updateEntityList {
		//e.EntityTree.SetSelection(objects)
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
				if body == nil {
					continue
				}
				p := body.Pos.Now
				if p[0]+body.Size.Render[0]*0.5 >= v1[0] && p[0]-body.Size.Render[0]*0.5 <= v2[0] &&
					p[1]+body.Size.Render[0]*0.5 >= v1[1] && p[1]-body.Size.Render[0]*0.5 <= v2[1] {
					if state.IndexOf(e.HoveringObjects, ibody) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, ibody)
					}
				}
			}
		}
	}
}

func (e *Editor) ResizeRenderer(w, h int) {
	e.Renderer = render.NewRenderer(e.DB)
	e.Renderer.ScreenWidth = w
	e.Renderer.ScreenHeight = h
	e.Renderer.Initialize()
}

func (e *Editor) MoveSurface(delta float64, floor bool, slope bool) {
	action := &actions.MoveSurface{IEditor: e, Delta: delta, Floor: floor, Slope: slope}
	e.NewAction(action)
	action.Act()
}

func (e *Editor) Alert(text string) {
	dlg := dialog.NewInformation("Foom Editor", text, e.Window)
	dlg.Show()
}

func (e *Editor) SetDialogLocation(dlg *dialog.FileDialog, target string) {
	if target == "" {
		target, _ = os.Getwd()
	}
	dlg.SetFileName(target)
	absPath, err := filepath.Abs(target)
	if err != nil {
		log.Printf("SetDialogLocation: error making absolute path from %v", target)
		absPath, _ = os.Getwd()
	}
	dir := filepath.Dir(absPath)
	uri := storage.NewFileURI(dir)
	lister, err := storage.ListerForURI(uri)
	if err != nil {
		log.Printf("SetDialogLocation: error making lister from %v", dir)
	} else {
		dlg.SetLocation(lister)
	}
}

func (e *Editor) ToolSelectSegment() {
	for _, s := range editor.SelectedObjects {
		switch target := s.(type) {
		case *concepts.EntityRef:
			if sector := core.SectorFromDb(target); sector != nil {
				editor.SelectObjects([]any{sector.Segments[0]}, true)
				break
			}
		case *core.Segment:
			editor.SelectObjects([]any{target.Next}, true)
			return
		}
	}
}
