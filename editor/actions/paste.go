// Pasteright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/driver/desktop"
)

// TODO: Be more flexible with what we can delete/cut/copy/paste. Should be
// possible to only grab a "hi" part of a segment, for example. For now,
// just grab EntityRefs
type Paste struct {
	state.IEditor

	Selected       []*core.Selectable
	CopiedToPasted map[uint64]uint64
	ClipboardData  string
}

func (a *Paste) Act() {
	a.ClipboardData = a.IEditor.Content()

	fmt.Printf("%v\n", a.ClipboardData)

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
			}
			selectable := core.SelectableFromEntityRef(c.Ref())
			selectable.AddToList(&a.Selected)
		}
	}

	// We need to wire up:
	// pasted materials to surfaces
	// pasted bodies to sectors
	// pasted internal segments to sectors
	for copiedEntity, pastedEntity := range a.CopiedToPasted {
		ref := db.EntityRef(pastedEntity)
		if seg := core.InternalSegmentFromDb(ref); seg != nil && seg.SectorEntityRef != nil {
			// If the parent was copied as well, switch to the pasted parent.
			if pastedParent, ok := a.CopiedToPasted[seg.SectorEntityRef.Entity]; ok {
				seg.SectorEntityRef = db.EntityRef(pastedParent)
				seg.Sector().InternalSegments[pastedEntity] = seg.Ref()
				delete(seg.Sector().InternalSegments, copiedEntity)
			}
		}
	}

	// Test
	m := concepts.IdentityMatrix2.Translate(&concepts.Vector2{100, 100})
	for _, s := range a.Selected {
		s.Transform(m)
	}
	a.State().Lock.Unlock()
	a.ActionFinished(false, true, true)
}
func (a *Paste) Cancel()                             {}
func (a *Paste) Frame()                              {}
func (a *Paste) OnMouseDown(evt *desktop.MouseEvent) {}
func (a *Paste) OnMouseMove()                        {}
func (a *Paste) OnMouseUp()                          {}

func (a *Paste) Undo() {
}
func (a *Paste) Redo() {

}
