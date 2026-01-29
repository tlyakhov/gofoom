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

	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/actions"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/loov/hrtime"
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

var eventKeyMap = map[string]dynamic.EventID{
	"W":      controllers.EventIdForward,
	"S":      controllers.EventIdBack,
	"E":      controllers.EventIdRight,
	"Q":      controllers.EventIdLeft,
	"A":      controllers.EventIdTurnLeft,
	"D":      controllers.EventIdTurnRight,
	"Return": controllers.EventIdPrimaryAction,
	"F":      controllers.EventIdSecondaryAction,
	"Space":  controllers.EventIdUp,
	"C":      controllers.EventIdDown,
}

type EditorWidgets struct {
	App          fyne.App
	Window       fyne.Window
	LabelStatus  *widget.Label
	PropertyGrid *fyne.Container
	GameWidget   *GameWidget
	MapWidget    *MapWidget
}

type Editor struct {
	state.EditorState
	// What we're editing.

	MapViewGrid
	EditorWidgets
	EditorMenu
	EntityList
	properties.Grid

	lastUndo int64

	// Game View state
	Renderer       *render.Renderer
	MapViewSurface *image.RGBA

	entityIconCache *xsync.MapOf[ecs.Entity, entityIconCacheItem]
	noTextureImage  image.Image
	lightImage      image.Image
	bodyImage       image.Image
	soundImage      image.Image
}

type entityIconCacheItem struct {
	image.Image
	LastUpdated int64
}

func (e *Editor) State() *state.EditorState {
	return &e.EditorState
}

func NewEditor() *Editor {
	e := &Editor{
		EditorState: state.EditorState{
			EditorSnapshot: state.EditorSnapshot{
				Scale:     1.0,
				Step:      10,
				GridB:     concepts.Vector2{1, 0},
				Selection: selection.NewSelection(),
			},
			// TODO: Save/Load these preferences
			Modified:                  false,
			BodiesVisible:             true,
			SectorTypesVisible:        false,
			ComponentNamesVisible:     true,
			DisabledPropertiesVisible: false,
			HoveringSelection:         selection.NewSelection(),
			KeysDown:                  make(containers.Set[fyne.KeyName]),
		},
		MapViewGrid:     MapViewGrid{Visible: true, Snap: true},
		entityIconCache: xsync.NewMapOf[ecs.Entity, entityIconCacheItem](),
	}
	e.Grid.IEditor = e
	e.Grid.MaterialSampler.Ray = &concepts.Ray{}
	e.ResizeRenderer(320, 240)
	e.MapViewGrid.Current = &e.EditorState.EditorSnapshot

	return e
}

func (e *Editor) OnStarted() {
	// This is used whenever we don't have a texture for something
	img := canvas.NewImageFromResource(theme.QuestionIcon())
	img.Resize(fyne.NewSquareSize(64))
	img.Refresh()
	editor.noTextureImage = img.Image

	img = canvas.NewImageFromResource(theme.RadioButtonCheckedIcon())
	img.Resize(fyne.NewSquareSize(64))
	img.Refresh()
	editor.lightImage = img.Image

	img = canvas.NewImageFromResource(theme.AccountIcon())
	img.Resize(fyne.NewSquareSize(64))
	img.Refresh()
	editor.bodyImage = img.Image

	img = canvas.NewImageFromResource(theme.VolumeUpIcon())
	img.Resize(fyne.NewSquareSize(64))
	img.Refresh()
	editor.soundImage = img.Image
}

func (e *Editor) Content() string {
	return e.App.Clipboard().Content()
}

func (e *Editor) SetContent(c string) {
	e.App.Clipboard().SetContent(c)
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

func (e *Editor) NewFrame() {

}

func (e *Editor) Integrate() {
	if player := e.Renderer.Player; player != nil {
		editor.GameInputLock.Lock()
		for key, eid := range eventKeyMap {
			if e.GameWidget.KeyMap.Contains(fyne.KeyName(key)) {
				ecs.Simulation.NewEvent(eid, &controllers.EntityEventParams{Entity: player.Entity})
			}
		}
		editor.GameInputLock.Unlock()
	}

	ecs.ActAllControllers(ecs.ControllerFrame)
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
	e.GameInputLock.Lock()
	defer e.GameInputLock.Unlock()
	defer e.SelectObjects(true)
	defer e.Lock.Unlock()
	e.OpenFile = filename
	e.Modified = false
	e.UpdateTitle()
	ecs.Initialize()
	err := ecs.Load(e.OpenFile)
	if err != nil {
		e.Alert(fmt.Sprintf("Error loading world: %v", err))
		return
	}
	controllers.RespawnAll()
	controllers.CreateFont(constants.DefaultFontPath, "Default Font")
	e.EntityList.ReIndex()
	ecs.Simulation.NewFrame = e.NewFrame
	ecs.Simulation.Integrate = e.Integrate
	ecs.Simulation.Render = e.GameWidget.Draw
	// TODO: this is a clunky way to set editor to paused, fix it.
	ecs.Simulation.EditorPaused = false
	e.BehaviorsPause.Menu.Action()
	e.entityIconCache.Clear()
	if e.Renderer != nil {
		e.Renderer.Initialize()
	}
}

func (e *Editor) Test() {
	e.Lock.Lock()
	defer e.SelectObjects(true)
	defer e.Lock.Unlock()

	e.Modified = false
	e.UpdateTitle()

	ecs.Initialize()
	controllers.CreateTestWorld2()
	e.EntityList.ReIndex()
	ecs.Simulation.NewFrame = e.NewFrame
	ecs.Simulation.Integrate = e.Integrate
	ecs.Simulation.Render = e.GameWidget.Draw
	e.entityIconCache.Clear()
	if e.Renderer != nil {
		e.Renderer.Initialize()
	}
}

func (e *Editor) autoPortal() {
	defer concepts.ExecutionDuration(concepts.ExecutionTrack("AutoPortal"))
	//e.Lock.Lock()
	controllers.AutoPortal()
	//e.Lock.Unlock()
}

func (e *Editor) refreshProperties() {
	// Execute UI updates on the main thread to prevent deadlocks when called from a background goroutine (e.g. SelectObjects)
	sel := e.Selection
	fyne.Do(func() {
		defer concepts.ExecutionDuration(concepts.ExecutionTrack("refreshProperties"))
		e.Grid.Refresh(sel)
		e.EntityList.Update()
	})
}

func (e *Editor) Snapshot(worldState bool) state.EditorSnapshot {
	result := e.EditorSnapshot

	if worldState {
		result.Snapshot = ecs.SaveSnapshot(true)
	}
	return result
}

func (e *Editor) ActionFinished(canceled, refreshProperties, autoPortal bool) {
	e.UpdateTitle()
	if autoPortal {
		e.autoPortal()
	}
	if refreshProperties {
		e.refreshProperties()
		for _, s := range e.Selection.Exact {
			e.EntityList.ReIndexComponents(s.Entity)
		}
	}
	e.SetMapCursor(desktop.DefaultCursor)
	e.CurrentAction = nil
	go e.UseTool()
}

func (e *Editor) Act(a state.Actionable) {
	e.Lock.Lock()
	defer e.Lock.Unlock()
	now := hrtime.Now().Milliseconds()

	// Wait at least 5sec between undo states for better performance.
	if now-e.lastUndo > 5*1000 {
		// TODO: Be smarter about when to snapshot, particularly avoid ECS
		// snapshots when actions don't modify the world. Also, for orthogonal
		// EditorState changes (for example, a pan followed by selection), we
		// should merge undo states somehow. Maybe add some kind of Flags enum
		// that actions could categorize themselves into, and sets of
		// consecutive actions of different flags could be merged.
		e.UndoHistory = append(e.UndoHistory, e.Snapshot(a.AffectsWorld()))
		if len(e.UndoHistory) > 100 {
			// TODO: Make ring buffer
			e.UndoHistory = e.UndoHistory[(len(e.UndoHistory) - 100):]
		}
	}
	e.RedoHistory = []state.EditorSnapshot{}
	e.CurrentAction = a
	a.Activate()
}

func (e *Editor) UseTool() {
	switch e.Tool {
	case state.ToolSplitSegment:
		e.Act(&actions.SplitSegment{
			Place: actions.Place{
				Action: state.Action{IEditor: e},
			}})
	case state.ToolSplitSector:
		e.Act(&actions.SplitSector{Place: actions.Place{Action: state.Action{IEditor: e}}})
	case state.ToolAddSector:
		s := &core.Sector{}
		s.Construct(nil)
		s.Bottom.Surface.Material = controllers.DefaultMaterial()
		s.Top.Surface.Material = controllers.DefaultMaterial()
		a := &actions.AddSector{}
		a.AddEntity.IEditor = e
		a.AddEntity.Components = []ecs.Component{s}
		e.Act(a)
	case state.ToolAddInternalSegment:
		seg := &core.InternalSegment{}
		seg.Construct(nil)
		a := &actions.AddInternalSegment{}
		a.AddEntity.IEditor = e
		a.AddEntity.Components = []ecs.Component{seg}
		e.Act(a)
	case state.ToolAddBody:
		body := &core.Body{}
		body.Construct(nil)
		e.Act(&actions.AddEntity{
			Place:      actions.Place{Action: state.Action{IEditor: e}},
			Components: []ecs.Component{body},
		})
	case state.ToolAlignGrid:
		e.Act(&actions.AlignGrid{Place: actions.Place{Action: state.Action{IEditor: e}}})
	case state.ToolPathDebug:
		e.Act(&actions.PathDebug{
			Place: actions.Place{
				Action: state.Action{IEditor: e},
			}})
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

		img := &materials.Image{}
		img.Construct(nil)
		img.Source = uc.URI().Path()
		img.Load()
		shader := &materials.Shader{}
		stage := &materials.ShaderStage{}
		stage.Construct(nil)
		shader.Stages = append(shader.Stages, stage)
		named := &ecs.Named{}
		named.Construct(nil)
		named.Name = "Shader " + path.Base(img.Source)

		a := &actions.AddEntity{}
		a.IEditor = e
		a.Components = []ecs.Component{img, shader, named}
		e.Act(a)

		stage.Material = a.Entity

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

func (e *Editor) UndoOrRedo(redo bool) {
	if m, ok := e.CurrentAction.(state.Cancelable); ok {
		m.Cancel()
	}
	e.Lock.Lock()
	defer e.Lock.Unlock()
	forward := &e.RedoHistory
	back := &e.UndoHistory
	if redo {
		forward, back = back, forward
	}
	index := len(*back) - 1
	if index < 0 {
		return
	}
	snapshot := (*back)[index]
	*back = (*back)[:index]
	// TODO: Remember whether the world state has been changed across actions
	*forward = append(*forward, e.Snapshot(true))
	if snapshot.Snapshot != nil {
		ecs.LoadSnapshot(snapshot.Snapshot)
	}
	e.Renderer.RefreshFont()
	e.Renderer.RefreshPlayer()
	prev := e.EditorSnapshot
	e.EditorSnapshot = snapshot
	if snapshot.Selection != nil {
		e.SetSelection(true, selection.NewSelectionClone(snapshot.Selection))
	} else {
		e.SetSelection(true, selection.NewSelection())
	}
	if prev.Tool != snapshot.Tool {
		e.SwitchTool(snapshot.Tool)
	}
	e.refreshProperties()
}

func (e *Editor) SelectObjects(updateEntityList bool, s ...*selection.Selectable) {
	editor.Lock.Lock()
	defer editor.Lock.Unlock()
	e.Selection.Clear()
	e.Selection.Add(s...)
	e.SetSelection(updateEntityList, e.Selection)
}

func (e *Editor) SetSelection(updateEntityList bool, s *selection.Selection) {
	e.Selection = s
	e.refreshProperties()
	if updateEntityList {
		e.EntityList.Select(e.Selection)
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

	e.HoveringSelection.Clear()

	// TODO: Make a "sector selection" mode where click-dragging selects sectors
	// rather than just segments.
	colSector := ecs.ArenaFor[core.Sector](core.SectorCID)
	for i := range colSector.Cap() {
		sector := colSector.Value(i)

		if sector == nil {
			continue
		}

		for _, segment := range sector.Segments {
			if editor.Selecting() {
				if segment.P.Render[0] >= v1[0] && segment.P.Render[1] >= v1[1] && segment.P.Render[0] <= v2[0] && segment.P.Render[1] <= v2[1] {
					e.HoveringSelection.Add(selection.SelectableFromSegment(segment))
				}
				/*if segment.AABBIntersect(v1[0], v1[1], v2[0], v2[1]) {
					if state.IndexOf(e.HoveringObjects, segment) == -1 {
						e.HoveringObjects = append(e.HoveringObjects, segment)
					}
				}*/
			} else {
				if e.MouseWorld.Sub(&segment.P.Render).Length() < state.SegmentSelectionEpsilon {
					e.HoveringSelection.Add(selection.SelectableFromSegment(segment))
				}
				if segment.DistanceToPoint(&e.MouseWorld) < state.SegmentSelectionEpsilon {
					e.HoveringSelection.Add(selection.SelectableFromSegment(segment))
				}
			}
		}
	}

	colSeg := ecs.ArenaFor[core.InternalSegment](core.InternalSegmentCID)
	for i := range colSeg.Cap() {
		seg := colSeg.Value(i)
		if seg == nil {
			continue
		}
		if editor.Selecting() {
			a := (seg.A[0] >= v1[0] && seg.A[1] >= v1[1] && seg.A[0] <= v2[0] && seg.A[1] <= v2[1])
			b := (seg.B[0] >= v1[0] && seg.B[1] >= v1[1] && seg.B[0] <= v2[0] && seg.B[1] <= v2[1])
			if a && b {
				e.HoveringSelection.Add(selection.SelectableFromInternalSegment(seg))
			} else if a {
				e.HoveringSelection.Add(selection.SelectableFromInternalSegmentA(seg))
			} else if b {
				e.HoveringSelection.Add(selection.SelectableFromInternalSegmentB(seg))
			}
		} else {
			if e.MouseWorld.Sub(seg.A).Length() < state.SegmentSelectionEpsilon {
				e.HoveringSelection.Add(selection.SelectableFromInternalSegmentA(seg))
			}
			if e.MouseWorld.Sub(seg.B).Length() < state.SegmentSelectionEpsilon {
				e.HoveringSelection.Add(selection.SelectableFromInternalSegmentB(seg))
			}
			if seg.DistanceToPoint(&e.MouseWorld) < state.SegmentSelectionEpsilon {
				e.HoveringSelection.Add(selection.SelectableFromInternalSegment(seg))
			}
		}
	}

	colWaypoint := ecs.ArenaFor[behaviors.ActionWaypoint](behaviors.ActionWaypointCID)
	for i := range colWaypoint.Cap() {
		aw := colWaypoint.Value(i)

		if aw == nil {
			continue
		}

		if editor.Selecting() {
			if aw.P[0] >= v1[0] && aw.P[1] >= v1[1] && aw.P[0] <= v2[0] && aw.P[1] <= v2[1] {
				e.HoveringSelection.Add(selection.SelectableFromActionWaypoint(aw))
			}
		} else {
			if e.MouseWorld.Sub(aw.P.To2D()).Length() < state.SegmentSelectionEpsilon {
				e.HoveringSelection.Add(selection.SelectableFromActionWaypoint(aw))
			}
			/*if segment.DistanceToPoint(&e.MouseWorld) < state.SegmentSelectionEpsilon {
				e.HoveringObjects.Add(core.SelectableFromPathSegment(segment))
			}*/
		}
	}

	colBody := ecs.ArenaFor[core.Body](core.BodyCID)
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
				e.HoveringSelection.Add(selection.SelectableFromBody(body))
			}
		} else {
			if p[0]+size > e.MouseWorld[0] && p[0]-size < e.MouseWorld[0] &&
				p[1]+size > e.MouseWorld[1] && p[1]-size < e.MouseWorld[1] {
				e.HoveringSelection.Add(selection.SelectableFromBody(body))
			}
		}
	}
}

func (e *Editor) ResizeRenderer(w, h int) {
	if e.Renderer == nil {
		e.Renderer = render.NewRenderer()
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
	for _, s := range editor.Selection.Grouped {
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

func (e *Editor) FocusedShortcut(s fyne.Shortcut) {
	if focused, ok := e.Window.Canvas().Focused().(fyne.Shortcutable); ok {
		focused.TypedShortcut(s)
	}
}
