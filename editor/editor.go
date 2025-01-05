// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/actions"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/puzpuzpuz/xsync/v3"

	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
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

	entityIconCache *xsync.MapOf[ecs.Entity, entityIconCacheItem]
	noTextureImage  image.Image
}

type entityIconCacheItem struct {
	image.Image
	LastUpdated int64
}

func (e *Editor) State() *state.Edit {
	return &e.Edit
}

func NewEditor() *Editor {
	e := &Editor{
		Edit: state.Edit{
			ECS: ecs.NewECS(),
			MapView: state.MapView{
				Scale: 1.0,
				Step:  10,
				GridB: concepts.Vector2{1, 0},
			},
			Modified:              false,
			BodiesVisible:         true,
			SectorTypesVisible:    false,
			ComponentNamesVisible: true,
			HoveringObjects:       selection.NewSelection(),
			SelectedObjects:       selection.NewSelection(),
			KeysDown:              make(containers.Set[fyne.KeyName]),
		},
		MapViewGrid:     MapViewGrid{Visible: true},
		entityIconCache: xsync.NewMapOf[ecs.Entity, entityIconCacheItem](),
	}
	e.Grid.IEditor = e
	e.Grid.MaterialSampler.Ray = &render.Ray{}
	e.ResizeRenderer(320, 240)
	e.MapViewGrid.Current = &e.Edit.MapView

	return e
}

func (e *Editor) OnStarted() {
	// This is used whenever we don't have a texture for something
	img := canvas.NewImageFromResource(theme.QuestionIcon())
	img.Resize(fyne.NewSquareSize(64))
	img.Refresh()
	editor.noTextureImage = img.Image
}

func (e *Editor) Content() string {
	return e.Window.Clipboard().Content()
}

func (e *Editor) SetContent(c string) {
	e.Window.Clipboard().SetContent(c)
}

func (e *Editor) ScreenToWorld(p *concepts.Vector2) *concepts.Vector2 {
	return p.Sub(e.Size.Mul(0.5)).MulSelf(1.0 / e.Scale).AddSelf(&e.Pos)
}

func (e *Editor) WorldToScreen(p *concepts.Vector2) *concepts.Vector2 {
	return &concepts.Vector2{
		(p[0]-e.Pos[0])*e.Scale + e.Size[0]*0.5,
		(p[1]-e.Pos[1])*e.Scale + e.Size[1]*0.5,
	}
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
	e.Lock.Lock()
	defer e.Lock.Unlock()

	var text string

	if m, ok := e.CurrentAction.(state.Cancelable); ok {
		text = m.Status()
	} else {
		text = "Left mouse down to transform selection. Right click/drag to select. Middle-drag to pan. Mouse-wheel to zoom."
	}

	if e.MousePressed {
		text += " Dragging "
		text += e.WorldGrid(&e.MouseDownWorld).StringHuman() + " -> " + e.WorldGrid(&e.MouseWorld).StringHuman()
		dist := e.WorldGrid(&e.MouseDownWorld).Sub(e.WorldGrid(&e.MouseWorld)).Length()
		text += " Length: " + strconv.FormatFloat(dist, 'f', 2, 64)
	} else {
		text += " Cursor: " + e.WorldGrid(&e.MouseWorld).StringHuman()
	}
	/*list := ""
	for _, s := range e.HoveringObjects.Exact {
		if len(list) > 0 {
			list += ", "
		}
		switch s.Type {
		case core.SelectableSector:
			list += s.Sector.String()
		case core.SelectableInternalSegmentA:
			fallthrough
		case core.SelectableInternalSegmentB:
			fallthrough
		case core.SelectableInternalSegment:
			list += s.InternalSegment.String()
		case core.SelectableBody:
			list += s.Body.String()
		case core.SelectableSectorSegment:
			list += s.SectorSegment.P.String()
		}

		if len(list) > 300 {
			list += "..."
			break
		}
	}
	text = list + " ( " + text + " )"*/
	e.LabelStatus.SetText(text)
}

func (e *Editor) Integrate() {
	editor.Lock.Lock()
	defer editor.Lock.Unlock()
	player := e.Renderer.Player
	if player == nil {
		return
	}
	playerBody := core.GetBody(player.ECS, player.Entity)
	playerMobile := core.GetMobile(player.ECS, player.Entity)

	if e.GameWidget.KeyMap.Contains("W") {
		controllers.MovePlayer(e.ECS, player.Entity, playerBody.Angle.Now)
	}
	if e.GameWidget.KeyMap.Contains("S") {
		controllers.MovePlayer(e.ECS, player.Entity, playerBody.Angle.Now+180.0)
	}
	if e.GameWidget.KeyMap.Contains("E") {
		controllers.MovePlayer(e.ECS, player.Entity, playerBody.Angle.Now+90.0)
	}
	if e.GameWidget.KeyMap.Contains("Q") {
		controllers.MovePlayer(e.ECS, player.Entity, playerBody.Angle.Now+270.0)
	}
	if e.GameWidget.KeyMap.Contains("A") {
		playerBody.Angle.Now -= constants.PlayerTurnSpeed * constants.TimeStepS
		playerBody.Angle.Now = concepts.NormalizeAngle(playerBody.Angle.Now)
	}
	if e.GameWidget.KeyMap.Contains("D") {
		playerBody.Angle.Now += constants.PlayerTurnSpeed * constants.TimeStepS
		playerBody.Angle.Now = concepts.NormalizeAngle(playerBody.Angle.Now)
	}

	player.ActionPressed = e.GameWidget.KeyMap.Contains("F")

	if e.GameWidget.KeyMap.Contains("Space") {
		if behaviors.GetUnderwater(player.ECS, playerBody.SectorEntity) != nil {
			playerMobile.Force[2] += constants.PlayerSwimStrength
		} else if playerBody.OnGround {
			playerMobile.Force[2] += constants.PlayerJumpForce
			playerBody.OnGround = false
		}
	}
	if e.GameWidget.KeyMap.Contains("C") {
		if behaviors.GetUnderwater(player.ECS, playerBody.SectorEntity) != nil {
			playerMobile.Force[2] -= constants.PlayerSwimStrength
		} else {
			player.Crouching = true
		}
	} else {
		player.Crouching = false
	}

	e.ECS.ActAllControllers(ecs.ControllerAlways)
	e.GatherHoveringObjects()
}

// TODO: This should be an action
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
	e.Modified = true
	e.UpdateTitle()
}

func (e *Editor) Load(filename string) {
	e.Lock.Lock()
	defer e.Lock.Unlock()
	e.OpenFile = filename
	e.Modified = false
	e.UpdateTitle()
	db := ecs.NewECS()
	err := db.Load(e.OpenFile)
	if err != nil {
		e.Alert(fmt.Sprintf("Error loading world: %v", err))
		return
	}
	controllers.Respawn(db, true)
	archetypes.CreateFont(db, "data/vga-font-8x8.png", "Default Font")
	db.Simulation.Integrate = e.Integrate
	db.Simulation.Render = e.GameWidget.Draw
	e.ECS = db
	if e.Renderer != nil {
		e.Renderer.ECS = db
		e.Renderer.Initialize()
	}
	e.SelectObjects(true)
}

func (e *Editor) Test() {
	e.Lock.Lock()
	defer e.Lock.Unlock()

	e.Modified = false
	e.UpdateTitle()

	e.ECS.Clear()
	controllers.CreateTestWorld2(e.ECS)
	e.ECS.Simulation.Integrate = e.Integrate
	e.ECS.Simulation.Render = e.GameWidget.Draw
	e.SelectObjects(true)
}

func (e *Editor) autoPortal() {
	defer concepts.ExecutionDuration(concepts.ExecutionTrack("AutoPortal"))
	//e.Lock.Lock()
	controllers.AutoPortal(e.ECS)
	//e.Lock.Unlock()
}

func (e *Editor) refreshProperties() {
	defer concepts.ExecutionDuration(concepts.ExecutionTrack("RefreshProperties"))
	e.Grid.Refresh(e.SelectedObjects)
	e.EntityList.Update()
}

func (e *Editor) ActionFinished(canceled, refreshProperties, autoPortal bool) {
	e.UpdateTitle()
	if autoPortal {
		e.autoPortal()
	}
	if !canceled {
		e.UndoHistory = append(e.UndoHistory, e.CurrentAction)
		if len(e.UndoHistory) > 100 {
			e.UndoHistory = e.UndoHistory[(len(e.UndoHistory) - 100):]
		}
		e.RedoHistory = []state.Actionable{}
	}
	if refreshProperties {
		e.refreshProperties()
	}
	e.SetMapCursor(desktop.DefaultCursor)
	e.CurrentAction = nil
	go e.UseTool()
}

func (e *Editor) Act(a state.Actionable) {
	e.Lock.Lock()
	defer e.Lock.Unlock()
	e.CurrentAction = a
	a.Activate()
}

func (e *Editor) UseTool() {
	switch e.Tool {
	case state.ToolSplitSegment:
		e.Act(&actions.SplitSegment{Place: actions.Place{IEditor: e}})
	case state.ToolSplitSector:
		e.Act(&actions.SplitSector{Place: actions.Place{IEditor: e}})
	case state.ToolAddSector:
		s := &core.Sector{}
		s.ComponentID = core.SectorCID
		s.Construct(nil)
		s.Bottom.Surface.Material = controllers.DefaultMaterial(e.ECS)
		s.Top.Surface.Material = controllers.DefaultMaterial(e.ECS)
		a := &actions.AddSector{}
		a.AddEntity.IEditor = e
		a.AddEntity.Components = []ecs.Attachable{s}
		e.Act(a)
	case state.ToolAddInternalSegment:
		seg := &core.InternalSegment{}
		seg.ComponentID = core.InternalSegmentCID
		seg.Construct(nil)
		a := &actions.AddInternalSegment{}
		a.AddEntity.IEditor = e
		a.AddEntity.Components = []ecs.Attachable{seg}
		e.Act(a)
	case state.ToolAddBody:
		body := &core.Body{}
		body.ComponentID = core.BodyCID
		body.Construct(nil)
		e.Act(&actions.AddEntity{
			Place:      actions.Place{IEditor: e},
			Components: []ecs.Attachable{body},
		})
	case state.ToolAlignGrid:
		e.Act(&actions.AlignGrid{IEditor: e})
	default:
		return
	}
}

func (e *Editor) NewShader() {
	dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil {
			e.Alert(fmt.Sprintf("Error loading file: %v", err))
			return
		}
		if uc == nil {
			return
		}
		// First, load the image
		eImg := e.ECS.NewEntity()
		img := e.ECS.NewAttachedComponent(eImg, materials.ImageCID).(*materials.Image)
		img.Source = uc.URI().Path()
		img.Load()
		a := &actions.AddEntity{Place: actions.Place{IEditor: e}, Entity: eImg, Components: e.ECS.AllComponents(eImg)}
		e.Act(a)
		e.CurrentAction.Activate()
		// Next set up the shader
		eShader := e.ECS.NewEntity()
		shader := e.ECS.NewAttachedComponent(eImg, materials.ShaderCID).(*materials.Shader)
		stage := &materials.ShaderStage{}
		stage.OnAttach(e.ECS)
		stage.Construct(nil)
		stage.Material = eImg
		shader.Stages = append(shader.Stages, stage)
		named := editor.ECS.NewAttachedComponent(eShader, ecs.NamedCID).(*ecs.Named)
		named.Name = "Shader " + path.Base(img.Source)
		a = &actions.AddEntity{Place: actions.Place{IEditor: e}, Entity: eShader, Components: e.ECS.AllComponents(eShader)}
		e.Act(a)
		e.CurrentAction.Activate()

	}, e.Window)

	dlg.Resize(fyne.NewSize(1000, 700))
	dlg.SetConfirmText("Load image file to use for shader")
	dlg.SetDismissText("Cancel")
	dlg.Show()
}

func (e *Editor) SwitchTool(tool state.EditorTool) {
	e.Tool = tool
	if m, ok := e.CurrentAction.(state.Cancelable); ok {
		m.Cancel()
	} else {
		e.UseTool()
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
	if e.CurrentAction != nil {
		// Don't undo if we're in the middle of a placing action
		if placeable, ok := e.CurrentAction.(actions.Placeable); ok && placeable.Placing() {
			return
		}
	}

	a.Undo()
	controllers.AutoPortal(e.ECS)
	e.refreshProperties()
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
	if e.CurrentAction != nil {
		// Don't redo if we're in the middle of a placing action
		if placeable, ok := e.CurrentAction.(actions.Placeable); ok && placeable.Placing() {
			return
		}
	}

	a.Redo()
	controllers.AutoPortal(e.ECS)
	e.refreshProperties()
	e.UndoHistory = append(e.UndoHistory, a)
}

func (e *Editor) SelectObjects(updateEntityList bool, s ...*selection.Selectable) {
	e.SelectedObjects.Clear()
	e.SelectedObjects.Add(s...)
	e.SetSelection(updateEntityList, e.SelectedObjects)
}

func (e *Editor) SetSelection(updateEntityList bool, s *selection.Selection) {
	e.SelectedObjects = s
	e.refreshProperties()
	if updateEntityList {
		e.EntityList.Select(e.SelectedObjects)
	}
}

func (e *Editor) Selecting() bool {
	_, ok := e.CurrentAction.(*actions.Select)
	return ok && e.MousePressed && e.MouseWorld.Dist(&e.MouseDownWorld) > state.SegmentSelectionEpsilon
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

	e.HoveringObjects.Clear()

	colSector := ecs.ColumnFor[core.Sector](e.ECS, core.SectorCID)
	for i := range colSector.Cap() {
		sector := colSector.Value(i)

		if sector == nil {
			continue
		}

		for _, segment := range sector.Segments {
			if editor.Selecting() {
				if segment.P[0] >= v1[0] && segment.P[1] >= v1[1] && segment.P[0] <= v2[0] && segment.P[1] <= v2[1] {
					e.HoveringObjects.Add(selection.SelectableFromSegment(segment))
				}
				/*if segment.AABBIntersect(v1[0], v1[1], v2[0], v2[1]) {
					if state.IndexOf(e.HoveringObjects, segment) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, segment)
					}
				}*/
			} else {
				if e.MouseWorld.Sub(&segment.P).Length() < state.SegmentSelectionEpsilon {
					e.HoveringObjects.Add(selection.SelectableFromSegment(segment))
				}
				if segment.DistanceToPoint(&e.MouseWorld) < state.SegmentSelectionEpsilon {
					e.HoveringObjects.Add(selection.SelectableFromSegment(segment))
				}
			}
		}
	}

	colSeg := ecs.ColumnFor[core.InternalSegment](e.ECS, core.InternalSegmentCID)
	for i := range colSeg.Cap() {
		seg := colSeg.Value(i)
		if seg == nil {
			continue
		}
		if editor.Selecting() {
			a := (seg.A[0] >= v1[0] && seg.A[1] >= v1[1] && seg.A[0] <= v2[0] && seg.A[1] <= v2[1])
			b := (seg.B[0] >= v1[0] && seg.B[1] >= v1[1] && seg.B[0] <= v2[0] && seg.B[1] <= v2[1])
			if a && b {
				e.HoveringObjects.Add(selection.SelectableFromInternalSegment(seg))
			} else if a {
				e.HoveringObjects.Add(selection.SelectableFromInternalSegmentA(seg))
			} else if b {
				e.HoveringObjects.Add(selection.SelectableFromInternalSegmentB(seg))
			}
		} else {
			if e.MouseWorld.Sub(seg.A).Length() < state.SegmentSelectionEpsilon {
				e.HoveringObjects.Add(selection.SelectableFromInternalSegmentA(seg))
			}
			if e.MouseWorld.Sub(seg.B).Length() < state.SegmentSelectionEpsilon {
				e.HoveringObjects.Add(selection.SelectableFromInternalSegmentB(seg))
			}
			if seg.DistanceToPoint(&e.MouseWorld) < state.SegmentSelectionEpsilon {
				e.HoveringObjects.Add(selection.SelectableFromInternalSegment(seg))
			}
		}
	}

	colWaypoint := ecs.ColumnFor[behaviors.ActionWaypoint](e.ECS, behaviors.ActionWaypointCID)
	for i := range colWaypoint.Cap() {
		aw := colWaypoint.Value(i)

		if aw == nil {
			continue
		}

		if editor.Selecting() {
			if aw.P[0] >= v1[0] && aw.P[1] >= v1[1] && aw.P[0] <= v2[0] && aw.P[1] <= v2[1] {
				e.HoveringObjects.Add(selection.SelectableFromActionWaypoint(aw))
			}
		} else {
			if e.MouseWorld.Sub(aw.P.To2D()).Length() < state.SegmentSelectionEpsilon {
				e.HoveringObjects.Add(selection.SelectableFromActionWaypoint(aw))
			}
			/*if segment.DistanceToPoint(&e.MouseWorld) < state.SegmentSelectionEpsilon {
				e.HoveringObjects.Add(core.SelectableFromPathSegment(segment))
			}*/
		}
	}

	colBody := ecs.ColumnFor[core.Body](e.ECS, core.BodyCID)
	for i := range colBody.Cap() {
		body := colBody.Value(i)
		if body == nil {
			continue
		}
		p := body.Pos.Now
		size := body.Size.Render[0]*0.5 + state.SegmentSelectionEpsilon
		if e.Selecting() {
			if p[0]+size >= v1[0] && p[0]-size <= v2[0] &&
				p[1]+size >= v1[1] && p[1]-size <= v2[1] {
				e.HoveringObjects.Add(selection.SelectableFromBody(body))
			}
		} else {
			if p[0]+size > e.MouseWorld[0] && p[0]-size < e.MouseWorld[0] &&
				p[1]+size > e.MouseWorld[1] && p[1]-size < e.MouseWorld[1] {
				e.HoveringObjects.Add(selection.SelectableFromBody(body))
			}
		}
	}
}

func (e *Editor) ResizeRenderer(w, h int) {
	if e.Renderer == nil {
		e.Renderer = render.NewRenderer(e.ECS)
	}
	e.Renderer.ScreenWidth = w
	e.Renderer.ScreenHeight = h
	e.Renderer.Initialize()
	e.Grid.MaterialSampler.Config = e.Renderer.Config
}

func (e *Editor) Alert(text string) {
	dlg := dialog.NewInformation("Foom Editor", text, e.Window)
	dlg.Show()
}

func (e *Editor) SetDialogLocation(dlg *dialog.FileDialog, target string) {
	if target == "" {
		target, _ = os.Getwd()
	}
	dlg.SetFileName(filepath.Base(target))
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
	for _, s := range editor.SelectedObjects.Grouped {
		switch s.Type {
		case selection.SelectableSector:
			editor.SelectObjects(true, selection.SelectableFromSegment(s.Sector.Segments[0]))
			return
		case selection.SelectableSectorSegment:
			editor.SelectObjects(true, selection.SelectableFromSegment(s.SectorSegment.Next))
			return
		}
	}
}
