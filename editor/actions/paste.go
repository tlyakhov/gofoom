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

	// Important note: the copied entity may have been modified or no longer
	// exist, since the user could have deleted/updated between cut/copying and
	// pasting.
	a.State().Lock.Lock()
	// Copied -> Pasted
	a.CopiedToPasted = make(map[uint64]uint64)
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
			c := db.LoadComponentWithoutAttaching(index, jsonComponent)

			if pastedEntity, ok := a.CopiedToPasted[copiedEntity]; ok {
				db.Attach(index, pastedEntity, c)
			} else {
				pastedEntity := db.NewEntity()
				db.Attach(index, pastedEntity, c)
				a.CopiedToPasted[copiedEntity] = pastedEntity
				pastedRef = c.Ref()
			}
		}
		if pastedRef != nil {
			selectable := core.SelectableFromEntityRef(pastedRef)
			selectable.AddToList(&a.Selected)
		}
	}

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

	// Save original positions
	a.Original = make([]concepts.Vector3, 0, len(a.Selected))
	for _, s := range a.Selected {
		a.Original = append(a.Original, s.SavePositions()...)
	}

	// Change selection
	a.SelectObjects(true, a.Selected...)
	a.State().Lock.Unlock()
}
func (a *Paste) Cancel() {}
func (a *Paste) Frame()  {}
func (a *Paste) OnMouseDown(evt *desktop.MouseEvent) {
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}
func (a *Paste) OnMouseMove() {
	a.Move.OnMouseMove()
}
func (a *Paste) OnMouseUp() {

}

func (a *Paste) Undo() {
}
func (a *Paste) Redo() {

}
