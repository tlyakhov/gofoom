package properties

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) fieldComponent(index int, field *state.PropertyGridField) {
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

	button, _ := gtk.ButtonNew()
	button.SetHExpand(true)
	button.SetLabel(fmt.Sprintf("Remove %v from [%v]", parentType.Elem().String(), entityList))
	button.Connect("clicked", func(_ *gtk.Button) {

		action := &actions.DeleteComponent{IEditor: g.IEditor, Components: make(map[uint64]concepts.Attachable)}
		for _, v := range field.Values {
			entity := v.Elem().Interface().(*concepts.EntityRef).Entity
			log.Printf("Detaching %v from %v", parentType.String(), entity)
			action.Components[entity] = field.Parent.(concepts.Attachable)
		}
		g.NewAction(action)
		action.Act()
		g.Container.GrabFocus()
	})
	g.Container.Attach(button, 2, index, 2, 1)
}
