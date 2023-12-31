package properties

import (
	"reflect"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) originalStrings(field *state.PropertyGridField) string {
	origValue := ""
	i := 0
	for _, v := range field.Unique {
		if i != 0 {
			origValue += ", "
		}
		origValue += v.Elem().String()
		i++
	}
	return origValue
}

func (g *Grid) fieldString(field *state.PropertyGridField) {
	origValue := g.originalStrings(field)

	entry := widget.NewEntry()
	entry.SetText(origValue)
	entry.OnSubmitted = func(text string) {
		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(text)}
		g.NewAction(action)
		action.Act()
		origValue = text
	}

	if exp, ok := field.Parent.(*core.Script); ok {
		label := widget.NewLabel("Compiled successfully")
		if exp.ErrorMessage != "" {
			label.Text = exp.ErrorMessage
			label.Importance = widget.DangerImportance
		} else {
			label.Importance = widget.SuccessImportance
		}
		g.GridWidget.Objects = append(g.GridWidget.Objects, container.NewVBox(entry, label))
	} else {
		g.GridWidget.Objects = append(g.GridWidget.Objects, entry)
	}
}
