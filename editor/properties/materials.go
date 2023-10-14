package properties

import (
	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) fieldMaterials(index int, field *pgField) {
	//var origValue uintptr
	if len(field.Values) > 1 {
		panic("Unexpectedly have multiple field values for a list of Materials.")
	}

	/*for _, v := range field.Values {
		origValue = v.Elem().Pointer()
	}*/

	button, _ := gtk.ButtonNew()
	button.SetHExpand(true)
	button.SetLabel("Add Material")
	button.Connect("clicked", func(_ *gtk.Button) {
		/*action := &actions.SetProperty{IEditor: g.IEditor, Fields: field.Values, ToSet: reflect.ValueOf(cb.GetActive())}
		g.NewAction(action)
		action.Act()
		//origValue = cb.GetActive()*/
		g.Container.GrabFocus()
	})
	g.Container.Attach(button, 2, index, 1, 1)
}
