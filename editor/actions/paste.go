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
	"sigs.k8s.io/yaml"
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

func (a *Paste) updateRelations(pastedEntity ecs.Entity) {
	ecs.RangeRelations(pastedEntity, func(r *ecs.Relation) bool {
		switch r.Type {
		case ecs.RelationOne:
			if pastedRelationEntity, ok := a.CopiedToPasted[r.One]; ok {
				r.One = pastedRelationEntity
				r.Update()
			}
		case ecs.RelationSet:
			for e := range r.Set {
				if pastedRelationEntity, ok := a.CopiedToPasted[e]; ok {
					r.Set.Delete(e)
					r.Set.Add(pastedRelationEntity)
				}
			}
			r.Update()
		case ecs.RelationSlice:
			for i, e := range r.Slice {
				if pastedRelationEntity, ok := a.CopiedToPasted[e]; ok {
					r.Slice[i] = pastedRelationEntity
				}
			}
			r.Update()
		case ecs.RelationTable:
			toDelete := make([]ecs.Entity, 0)
			for _, e := range r.Table {
				if e == 0 {
					continue
				}
				if pastedRelationEntity, ok := a.CopiedToPasted[e]; ok {
					toDelete = append(toDelete, e)
					r.Table.Set(pastedRelationEntity)
				}
			}
			for _, e := range toDelete {
				r.Table.Delete(e)
			}
			r.Update()
		}
		return true
	})
	if seg := core.GetInternalSegment(pastedEntity); seg != nil {
		seg.AttachToSectors()
	}
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
	for copiedEntityString, yamlData := range yamlEntities {
		copiedEntity, _ := ecs.ParseEntity(copiedEntityString)
		yamlEntity := yamlData.(map[string]any)
		if yamlEntity == nil {
			log.Printf("Paste.Activate: ECS YAML object element should be an object")
			continue
		}

		var pastedEntity ecs.Entity
		var ok bool
		for name, id := range ecs.Types().IDs {
			yamlData := yamlEntity[name]
			if yamlData == nil {
				continue
			}
			var toAttach ecs.Component
			switch yamlDataTyped := yamlData.(type) {
			case string: // It's an entity reference
				entityRef, err := ecs.ParseEntity(yamlDataTyped)
				if err != nil {
					log.Printf("Paste.Activate: Error parsing copied entity %v (component %v)", yamlDataTyped, name)
					continue
				}
				toAttach = ecs.GetComponent(entityRef, id)
			case map[string]any:
				toAttach = ecs.LoadComponentWithoutAttaching(id, yamlDataTyped)
			default:
				log.Printf("Paste.Activate: Unexpected type of copied component %v: %v", name, yamlDataTyped)
				continue
			}
			if pastedEntity, ok = a.CopiedToPasted[copiedEntity]; ok {
				ecs.Attach(id, pastedEntity, &toAttach)
			} else {
				pastedEntity = ecs.NewEntity()
				ecs.Attach(id, pastedEntity, &toAttach)
				a.CopiedToPasted[copiedEntity] = pastedEntity
			}
		}
		if pastedEntity != 0 {
			a.Selected.Add(selection.SelectableFromEntity(pastedEntity))
		}
	}

	for _, pastedEntity := range a.CopiedToPasted {
		a.updateRelations(pastedEntity)
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

func (a *Paste) MouseDown(evt *desktop.MouseEvent) {
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}
func (a *Paste) MouseMoved(evt *desktop.MouseEvent) {
	a.Delta = *a.State().MouseWorld.Sub(a.WorldGrid(a.Center.To2D()))
	a.Transform.Apply()
}
func (a *Paste) MouseUp(evt *desktop.MouseEvent) {

}

func (a *Paste) Undo() {
	for _, pasted := range a.CopiedToPasted {
		ecs.Delete(pasted)
	}
}

func (a *Paste) Redo() {
	a.apply()
	a.Transform.Apply()
}

func (a *Paste) Status() string {
	return "Click to place pasted entity/entities"
}
