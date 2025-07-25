// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"fmt"
	"log"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldComponent(field *state.PropertyGridField) {
	disable := true
	parentType := ""
	var parent ecs.Attachable
	var parentCID ecs.ComponentID
	allEntities := make(ecs.EntityTable, 0)
	selEntities := make(ecs.EntityTable, 0)
	for _, v := range field.Values {
		subTable := v.Interface().(ecs.EntityTable)
		for _, e := range subTable {
			if e != 0 {
				allEntities.Set(e)
			}
		}
		selEntities.Set(v.Entity)
		if !v.Entity.IsExternal() {
			disable = false
		}
		parent = v.Parent().(ecs.Attachable)
		parentCID = parent.ComponentID()
		parentType = ecs.Types().ArenaPlaceholders[parentCID].Type().Name()
	}

	label := widget.NewLabel("Entities: " + concepts.TruncateString(allEntities.String(), 20))

	removeButton := widget.NewButton("", nil)
	removeButton.Text = fmt.Sprintf("Remove %v from [%v]", parentType, concepts.TruncateString(selEntities.String(), 10))
	if len(removeButton.Text) > 32 {
		removeButton.Text = removeButton.Text[:32] + "..."
	}
	removeButton.Icon = theme.ContentRemoveIcon()
	removeButton.OnTapped = func() {
		action := &actions.UpdateLinks{
			Action:           state.Action{IEditor: g.IEditor},
			Entities:         selEntities,
			RemoveComponents: make(containers.Set[ecs.ComponentID]),
		}
		action.RemoveComponents.Add(parentCID)
		for _, e := range selEntities {
			log.Printf("Detaching %v from %v", parentType, e)
		}
		g.Act(action)
		g.Focus(g.GridWidget)
	}

	entitiesEntry := widget.NewEntry()
	entitiesEntry.Text = ""
	addButton := widget.NewButton("", nil)
	addButton.Text = concepts.TruncateString(fmt.Sprintf("Link %v...", parentType), 32)
	addButton.Icon = theme.ContentAddIcon()
	addButton.OnTapped = func() {
		dialog.ShowForm(addButton.Text, "Add", "Cancel", []*widget.FormItem{
			{Text: "Entities", Widget: entitiesEntry},
		}, func(b bool) {
			if !b {
				return
			}
			action := &actions.UpdateLinks{
				Action:        state.Action{IEditor: g.IEditor},
				Entities:      ecs.ParseEntityTable(entitiesEntry.Text, true),
				AddComponents: make(ecs.ComponentTable, 0),
			}
			action.AddComponents.Set(parent)
			for _, e := range action.Entities {
				if e != 0 {
					log.Printf("Attaching %v to %v", parentType, e)
				}
			}
			g.Act(action)
			g.Focus(g.GridWidget)
		}, g.GridWindow)
	}

	if disable {
		addButton.Disable()
		removeButton.Disable()
	} else {
		addButton.Enable()
		removeButton.Enable()
	}

	c := gridAddOrUpdateWidgetAtIndex[*fyne.Container](g)
	c.Layout = layout.NewVBoxLayout()
	if parent.MultiAttachable() {
		c.Objects = []fyne.CanvasObject{label, addButton, removeButton}
	} else {
		c.Objects = []fyne.CanvasObject{removeButton}
	}
}
