package main

import (
	"image/color"
	"strconv"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/image/colornames"
)

type EntityListColumnID int

//go:generate go run github.com/dmarkham/enumer -type=EntityListColumnID -json
const (
	elcEntity EntityListColumnID = iota
	elcDesc
	elcRank
	elcColor
)

type EntityList struct {
	state.IEditor

	Table     *widget.Table
	Container *fyne.Container

	BackingStore [][4]any
}

func (list *EntityList) tableLength() (rows int, cols int) {
	cols = 3
	rows = len(list.BackingStore)
	return
}

func (list *EntityList) tableUpdate(tci widget.TableCellID, template fyne.CanvasObject) {
	if list.BackingStore == nil { // || len(list.BackingStore) <= tci.Row {
		return
	}
	row := list.BackingStore[tci.Row]

	stack := template.(*fyne.Container)
	text := stack.Objects[0].(*canvas.Text)
	progress := stack.Objects[1].(*widget.ProgressBar)
	switch tci.Col {
	case int(elcEntity):
		progress.Hide()
		text.Color = row[elcColor].(color.Color)
		text.Text = strconv.Itoa(row[elcEntity].(int))
		text.Show()
		text.Refresh()
	case int(elcDesc):
		progress.Hide()
		text.Color = row[elcColor].(color.Color)
		text.Text = row[elcDesc].(string)
		if len(text.Text) > 30 {
			text.Text = text.Text[:30] + "..."
		}
		text.Show()
		text.Refresh()
	case int(elcRank):
		text.Hide()
		progress.SetValue(float64(row[elcRank].(int)) / 100)
		progress.Show()
		progress.Refresh()
	}
}

func (list *EntityList) Build() fyne.CanvasObject {
	list.Update()

	list.Table = widget.NewTableWithHeaders(list.tableLength, func() fyne.CanvasObject {
		text := canvas.NewText("Template", theme.ForegroundColor())
		progress := widget.NewProgressBar()
		return container.NewStack(text, progress)
	}, list.tableUpdate)

	list.Table.SetColumnWidth(int(elcEntity), 50)
	list.Table.SetColumnWidth(int(elcDesc), 280)
	list.Table.SetColumnWidth(int(elcRank), 100)
	list.Table.ShowHeaderColumn = false
	list.Table.CreateHeader = func() fyne.CanvasObject {
		return widget.NewButton("Template", func() {})
	}
	list.Table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		b := template.(*widget.Button)
		if id.Col == -1 {
			b.Hide()
			return
		}
		switch id.Col {
		case int(elcEntity):
			b.SetText("Entity")
		case int(elcDesc):
			b.SetText("Desc")
		case int(elcRank):
			b.SetText("Rank")
		}
		b.Importance = widget.MediumImportance
		b.OnTapped = func() {
		}
		b.Enable()
		b.Refresh()
	}

	button := widget.NewButtonWithIcon("Add Empty Entity", theme.ContentAddIcon(), func() {
		list.State().Lock.Lock()
		ref := editor.DB.NewEntityRef()
		editor.SelectObjects([]any{ref}, true)
		list.State().Lock.Unlock()
	})
	search := widget.NewEntry()
	search.SetPlaceHolder("Search for entity...")
	search.OnChanged = func(s string) {
		list.State().SearchTerms = s
		list.Update()
	}

	list.Container = container.NewBorder(container.NewVBox(button, search), nil, nil, nil, list.Table)
	return list.Container
}
func (list *EntityList) Update() {
	list.BackingStore = make([][4]any, 0)
	ec := list.State().DB.EntityComponents
	if ec == nil {
		return
	}

	searchValid := len(list.State().SearchTerms) > 0

	for entity, components := range ec {
		if components == nil {
			continue
		}
		rowColor := theme.ForegroundColor()
		parentDesc := ""
		rank := 0
		for index, c := range components {
			if c == nil {
				continue
			}
			desc := c.String()

			if searchValid {
				rank += fuzzy.RankMatchFold(list.State().SearchTerms, desc)
			}

			if len(parentDesc) > 0 {
				parentDesc += "|"
			}
			parentDesc += desc

			if index == core.BodyComponentIndex {
				rowColor = colornames.Lightblue
			} else if index == core.SectorComponentIndex {
				rowColor = colornames.Lightgreen
			} else if index == materials.ImageComponentIndex ||
				index == materials.SolidComponentIndex ||
				index == materials.ShaderComponentIndex {
				rowColor = colornames.Lightpink
			}
		}
		dispRank := 100
		if searchValid {
			dispRank = concepts.Max(concepts.Min(rank+50, 100), 0)
		}
		backingRow := [4]any{
			entity,
			parentDesc,
			dispRank,
			rowColor,
		}
		list.BackingStore = append(list.BackingStore, backingRow)
	}
	if list.Table != nil {
		list.Table.Refresh()
	}
}
