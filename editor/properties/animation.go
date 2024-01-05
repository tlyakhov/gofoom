package properties

import (
	"reflect"
	"slices"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldAnimation(field *state.PropertyGridField) {
	origValue := field.Values[0].Elem()
	if origValue.IsNil() {
		button := widget.NewButtonWithIcon("Add Animation", theme.ContentAddIcon(), func() {
			parentValue := reflect.ValueOf(field.Parent)
			m := parentValue.MethodByName("NewAnimation")
			newAnimation := m.Call(nil)[0]
			action := &actions.SetProperty{
				IEditor:           g.IEditor,
				PropertyGridField: field,
				ToSet:             newAnimation,
			}
			g.NewAction(action)
			action.Act()
		})
		g.GridWidget.Objects = append(g.GridWidget.Objects, button)
	} else {
		button := widget.NewButtonWithIcon("Remove Animation", theme.ContentClearIcon(), func() {
			action := &actions.SetProperty{
				IEditor:           g.IEditor,
				PropertyGridField: field,
				ToSet:             reflect.Zero(origValue.Type()),
			}
			g.NewAction(action)
			action.Act()
		})
		g.GridWidget.Objects = append(g.GridWidget.Objects, button)
	}
}

func (g *Grid) fieldTweeningFunc(field *state.PropertyGridField) {
	origValue := field.Values[0].Elem().Pointer()

	opts := make([]string, 0)
	for name := range concepts.TweeningFuncs {
		opts = append(opts, name)
	}
	slices.Sort(opts)
	optValues := make([]reflect.Value, 0)
	selectedIndex := 0
	for i, name := range opts {
		f := reflect.ValueOf(concepts.TweeningFuncs[name])
		optValues = append(optValues, f)
		if f.Pointer() == origValue {
			selectedIndex = i
		}
	}

	s := widget.NewSelect(opts, nil)
	s.SetSelectedIndex(selectedIndex)
	s.PlaceHolder = "Select tweening function"
	s.OnChanged = func(opt string) {
		action := &actions.SetProperty{
			IEditor:           g.IEditor,
			PropertyGridField: field,
			ToSet:             optValues[s.SelectedIndex()],
		}
		g.NewAction(action)
		action.Act()
	}

	g.GridWidget.Objects = append(g.GridWidget.Objects, s)
}
