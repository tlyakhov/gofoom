package properties

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldComponent(field *state.PropertyGridField) {
	// This will be a pointer
	parentType := reflect.TypeOf(field.Parent)
	entityList := ""
	for _, v := range field.Values {
		er := v.Elem().Interface().(*concepts.EntityRef)
		entity := er.Entity
		if len(entityList) > 0 {
			entityList += ", "
		}
		entityList += strconv.FormatUint(entity, 10)
	}

	bLabel := fmt.Sprintf("Remove %v from [%v]", parentType.Elem().String(), entityList)
	button := widget.NewButtonWithIcon(bLabel, theme.ContentRemoveIcon(), func() {
		action := &actions.DeleteComponent{IEditor: g.IEditor, Components: make(map[uint64]concepts.Attachable)}
		for _, v := range field.Values {
			entity := v.Elem().Interface().(*concepts.EntityRef).Entity
			log.Printf("Detaching %v from %v", parentType.String(), entity)
			action.Components[entity] = field.Parent.(concepts.Attachable)
		}
		g.NewAction(action)
		action.Act()
		g.Focus(g.GridWidget)

	})
	g.GridWidget.Objects = append(g.GridWidget.Objects, button)

}
