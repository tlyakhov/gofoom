// Pasteright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"log"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"

	"fyne.io/fyne/v2/driver/desktop"
	"gopkg.in/yaml.v3"
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

func (a *Paste) apply() {
	var parsed any
	err := yaml.Unmarshal([]byte(a.ClipboardData), &parsed)
	if err != nil {
		log.Printf("Paste.Activate: Clipboard data is not valid YAML set of entities: %v\n", err)
		return
	}

	var yamlEntities map[string]any
	var ok bool
	if yamlEntities, ok = parsed.(map[string]any); !ok || yamlEntities == nil {
		log.Printf("Paste.Activate: Clipboard YAML root must be an object")
		return
	}

	// Important note: the copied objects may have been modified or no longer
	// exist, since the user could have deleted/updated between cut/copying and
	// pasting.

	// Copied -> Pasted
	a.CopiedToPasted = make(map[ecs.Entity]ecs.Entity)
	a.Selected = selection.NewSelection()
	u := a.State().Universe
	for copiedEntityString, yamlData := range yamlEntities {
		copiedEntity, _ := ecs.ParseEntity(copiedEntityString)
		yamlEntity := yamlData.(map[string]any)
		if yamlEntity == nil {
			log.Printf("Paste.Activate: Universe YAML object element should be an object")
			continue
		}

		var pastedEntity ecs.Entity
		var ok bool
		for name, id := range ecs.Types().IDs {
			yamlData := yamlEntity[name]
			if yamlData == nil {
				continue
			}
			yamlComponent := yamlData.(map[string]any)
			c := u.LoadComponentWithoutAttaching(id, yamlComponent)

			if pastedEntity, ok = a.CopiedToPasted[copiedEntity]; ok {
				u.Attach(id, pastedEntity, &c)
			} else {
				pastedEntity = u.NewEntity()
				u.Attach(id, pastedEntity, &c)
				a.CopiedToPasted[copiedEntity] = pastedEntity
			}
		}
		if pastedEntity != 0 {
			a.Selected.Add(selection.SelectableFromEntity(u, pastedEntity))
		}
	}

	// We need to wire up:
	// pasted materials to surfaces
	// pasted bodies to sectors
	// pasted internal segments to sectors
	for _, pastedEntity := range a.CopiedToPasted {
		if seg := core.GetInternalSegment(u, pastedEntity); seg != nil {
			seg.AttachToSectors()
		}
		if body := core.GetBody(u, pastedEntity); body != nil {
			body.SectorEntity = 0
		}
		// TODO: materials
	}

	a.Selected.SavePositions()
	a.Center[0] = 0
	a.Center[1] = 0
	a.Center[2] = 0
	// Calculate the center of the selection
	for _, pos := range a.Selected.Positions {
		a.Center.AddSelf(pos)
	}
	a.Center.MulSelf(1.0 / float64(len(a.Selected.Positions)))

	// Change selection
	a.SetSelection(true, a.Selected)
}

func (a *Paste) Activate() {
	a.ClipboardData = a.IEditor.Content()
	a.apply()
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
	for _, pasted := range a.CopiedToPasted {
		a.State().Universe.Delete(pasted)
	}
}

func (a *Paste) Redo() {
	a.apply()
	a.Transform.Apply()
}

func (a *Paste) Status() string {
	return "Click to place pasted entity/entities"
}
