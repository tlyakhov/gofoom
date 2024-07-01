// Pasteright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"encoding/json"
	"log"
	"strconv"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/driver/desktop"
)

// TODO: Be more flexible with what we can delete/cut/copy/paste. Should be
// possible to only grab a "hi" part of a segment, for example. For now,
// just grab EntityRefs
type Paste struct {
	Move

	CopiedToPasted map[uint64]uint64
	ClipboardData  string
	Center         concepts.Vector3
}

func (a *Paste) Act() {
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
	a.CopiedToPasted = make(map[uint64]uint64)
	a.Selected = core.NewSelection()
	db := a.State().DB
	for copiedEntityString, jsonData := range jsonEntities {
		copiedEntity, _ := strconv.ParseUint(copiedEntityString, 10, 64)
		jsonEntity := jsonData.(map[string]any)
		if jsonEntity == nil {
			log.Printf("ECS JSON object element should be an object\n")
			continue
		}

		var pastedRef *concepts.EntityRef
		for name, index := range concepts.DbTypes().Indexes {
			jsonData := jsonEntity[name]
			if jsonData == nil {
				continue
			}
			jsonComponent := jsonData.(map[string]any)
			a.State().Lock.Lock()
			c := db.LoadComponentWithoutAttaching(index, jsonComponent)

			if pastedEntity, ok := a.CopiedToPasted[copiedEntity]; ok {
				db.Attach(index, pastedEntity, c)
			} else {
				pastedEntity := db.NewEntity()
				db.Attach(index, pastedEntity, c)
				a.CopiedToPasted[copiedEntity] = pastedEntity
				pastedRef = c.Ref()
			}
			a.State().Lock.Unlock()
		}
		if pastedRef != nil {
			a.Selected.Add(core.SelectableFromEntityRef(pastedRef))
		}
	}

	a.State().Lock.Lock()
	// We need to wire up:
	// pasted materials to surfaces
	// pasted bodies to sectors
	// pasted internal segments to sectors
	for _, pastedEntity := range a.CopiedToPasted {
		ref := db.EntityRef(pastedEntity)
		if seg := core.InternalSegmentFromDb(ref); seg != nil {
			seg.AttachToSectors()
		}
		if body := core.BodyFromDb(ref); body != nil {
			body.SectorEntityRef = nil
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
func (a *Paste) Frame()  {}
func (a *Paste) OnMouseDown(evt *desktop.MouseEvent) {
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}
func (a *Paste) OnMouseMove() {
	a.Delta = *a.State().MouseWorld.Sub(a.WorldGrid(a.Center.To2D()))
	a.Move.Act()
}
func (a *Paste) OnMouseUp() {

}

func (a *Paste) Undo() {
	// TODO: Implement
}
func (a *Paste) Redo() {
	// TODO: Implement
}
