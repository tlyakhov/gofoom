// Pasteright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"encoding/json"
	"log"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"

	"fyne.io/fyne/v2/driver/desktop"
)

// TODO: Be more flexible with what we can delete/cut/copy/paste. Should be
// possible to only grab a "hi" part of a segment, for example. For now,
// just grab EntityRefs
type Paste struct {
	Transform

	CopiedToPasted map[ecs.Entity]ecs.Entity
	ClipboardData  string
	Center         concepts.Vector3
}

func (a *Paste) Activate() {
	a.ClipboardData = a.IEditor.Content()

	var parsed any
	err := json.Unmarshal([]byte(a.ClipboardData), &parsed)
	if err != nil {
		log.Printf("Clipboard data is not valid JSON set of entities: %v\n", err)
		return
	}

	var jsonEntities map[string]any
	var ok bool
	if jsonEntities, ok = parsed.(map[string]any); !ok || jsonEntities == nil {
		log.Printf("Clipboard JSON root must be an object")
		return
	}

	// Important note: the copied objects may have been modified or no longer
	// exist, since the user could have deleted/updated between cut/copying and
	// pasting.

	// Copied -> Pasted
	a.CopiedToPasted = make(map[ecs.Entity]ecs.Entity)
	a.Selected = selection.NewSelection()
	db := a.State().ECS
	for copiedEntityString, jsonData := range jsonEntities {
		copiedEntity, _ := ecs.ParseEntity(copiedEntityString)
		jsonEntity := jsonData.(map[string]any)
		if jsonEntity == nil {
			log.Printf("ECS JSON object element should be an object\n")
			continue
		}

		var pastedEntity ecs.Entity
		var ok bool
		for name, id := range ecs.Types().IDs {
			jsonData := jsonEntity[name]
			if jsonData == nil {
				continue
			}
			jsonComponent := jsonData.(map[string]any)
			a.State().Lock.Lock()
			c := db.LoadComponentWithoutAttaching(id, jsonComponent)

			if pastedEntity, ok = a.CopiedToPasted[copiedEntity]; ok {
				db.Attach(id, pastedEntity, c)
			} else {
				pastedEntity = db.NewEntity()
				db.Attach(id, pastedEntity, c)
				a.CopiedToPasted[copiedEntity] = pastedEntity
			}
			a.State().Lock.Unlock()
		}
		if pastedEntity != 0 {
			a.Selected.Add(selection.SelectableFromEntity(db, pastedEntity))
		}
	}

	a.State().Lock.Lock()
	// We need to wire up:
	// pasted materials to surfaces
	// pasted bodies to sectors
	// pasted internal segments to sectors
	for _, pastedEntity := range a.CopiedToPasted {
		if seg := core.GetInternalSegment(db, pastedEntity); seg != nil {
			seg.AttachToSectors()
		}
		if body := core.GetBody(db, pastedEntity); body != nil {
			body.SectorEntity = 0
		}
		// TODO: materials
	}

	a.Selected.SavePositions()
	// Calculate the center of the selection
	for _, pos := range a.Selected.Positions {
		a.Center.AddSelf(pos)
	}
	a.Center.MulSelf(1.0 / float64(len(a.Selected.Positions)))

	a.State().Lock.Unlock()

	// Change selection
	a.SetSelection(true, a.Selected)
}
func (a *Paste) Cancel() {}

func (a *Paste) OnMouseDown(evt *desktop.MouseEvent) {
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}
func (a *Paste) OnMouseMove() {
	a.Delta = *a.State().MouseWorld.Sub(a.WorldGrid(a.Center.To2D()))
	a.Transform.Apply()
}
func (a *Paste) OnMouseUp() {

}

func (a *Paste) Undo() {
	// TODO: Implement
}
func (a *Paste) Redo() {
	// TODO: Implement
}

func (a *Paste) Status() string {
	return "Click to place pasted entity/entities"
}
