// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"image/color"
	"log"
	"slices"
	"sort"
	"strings"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
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
	elcImage
	elcDesc
	elcRank
	elcColor
	elcNumColumns
)

type elSortDir int

const (
	elsdSortOff elSortDir = iota
	elsdSortAsc
	elsdSortDesc
)

type EntityList struct {
	state.IEditor

	Table     *widget.Table
	Container *fyne.Container

	BackingStore [][elcNumColumns]any
	Sorts        [elcNumColumns]elSortDir
}

func (list *EntityList) tableLength() (rows int, cols int) {
	cols = 4
	rows = len(list.BackingStore)
	return
}

func (list *EntityList) tableUpdate(tci widget.TableCellID, template fyne.CanvasObject) {
	if list.BackingStore == nil {
		return
	}
	row := list.BackingStore[tci.Row]

	stack := template.(*fyne.Container)
	text := stack.Objects[0].(*canvas.Text)
	progress := stack.Objects[1].(*widget.ProgressBar)
	img := stack.Objects[2].(*canvas.Image)

	switch tci.Col {
	case int(elcEntity):
		progress.Hide()
		img.Hide()
		text.Color = row[elcColor].(color.Color)
		e := ecs.Entity(row[elcEntity].(int))
		text.Text = e.ShortString()
		text.TextSize = theme.TextSize()
		text.Show()
		text.Refresh()
	case int(elcImage):
		progress.Hide()
		text.Hide()
		e := ecs.Entity(row[elcEntity].(int))
		img.Image = list.IEditor.EntityImage(e)
		img.Show()
		img.Refresh()
	case int(elcDesc):
		progress.Hide()
		img.Hide()
		text.Color = row[elcColor].(color.Color)
		text.Text = concepts.TruncateString(row[elcDesc].(string), 60)
		text.TextSize = 10
		text.Show()
		text.Refresh()
	case int(elcRank):
		img.Hide()
		text.Hide()
		progress.SetValue(float64(row[elcRank].(int)) / 100)
		progress.Show()
		progress.Refresh()
	}
}

func (list *EntityList) Build() fyne.CanvasObject {
	list.Sorts[0] = elsdSortAsc
	list.Update()

	list.Table = widget.NewTableWithHeaders(list.tableLength, func() fyne.CanvasObject {
		text := canvas.NewText("Template", theme.Color(theme.ColorNameForeground))
		progress := widget.NewProgressBar()
		img := canvas.NewImageFromImage(nil)
		img.ScaleMode = canvas.ImageScaleSmooth
		img.FillMode = canvas.ImageFillContain
		//	img.SetMinSize(fyne.NewSquareSize(64))
		return container.NewStack(text, progress, img)
	}, list.tableUpdate)

	list.Table.SetColumnWidth(int(elcEntity), 70)
	list.Table.SetColumnWidth(int(elcImage), 64)
	list.Table.SetColumnWidth(int(elcDesc), 240)
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
		case int(elcImage):
			b.SetText("")
		case int(elcDesc):
			b.SetText("Desc")
		case int(elcRank):
			b.SetText("Rank")
		}
		switch list.Sorts[id.Col] {
		case elsdSortAsc:
			b.Icon = theme.MoveUpIcon()
		case elsdSortDesc:
			b.Icon = theme.MoveDownIcon()
		default:
			b.Icon = nil
		}
		b.Importance = widget.MediumImportance
		b.OnTapped = func() {
			list.AdvanceSort(id.Col)
			list.applySort()
		}
		b.Enable()
		b.Refresh()
	}

	button := widget.NewButtonWithIcon("Add Empty Entity", theme.ContentAddIcon(), func() {
		list.State().Lock.Lock()
		editor.SelectObjects(true, selection.SelectableFromEntity(editor.Universe, editor.Universe.NewEntity()))
		list.State().Lock.Unlock()
	})
	search := widget.NewEntry()
	search.ActionItem = widget.NewIcon(theme.SearchIcon())
	search.SetPlaceHolder("Search for entity...")
	search.OnChanged = func(s string) {
		list.State().SearchTerms = s
		list.SetSort(2, elsdSortDesc)
		list.Update()
	}

	list.Table.OnSelected = func(id widget.TableCellID) {
		if list.BackingStore == nil || id.Row < 0 || id.Row >= len(list.BackingStore) {
			return
		}
		entity := list.BackingStore[id.Row][0].(int)
		s := selection.SelectableFromEntity(list.State().Universe, ecs.Entity(entity))
		if !editor.SelectedObjects.Contains(s) {
			editor.SelectObjects(false, s)
		}
		log.Printf("select: %v", entity)
	}
	list.Table.OnUnselected = func(id widget.TableCellID) {
		if list.BackingStore == nil || id.Row < 0 || id.Row >= len(list.BackingStore) {
			return
		}
		entity := list.BackingStore[id.Row][0].(int)
		log.Printf("unselect: %v", entity)
	}

	list.Container = container.NewBorder(container.NewVBox(button, search), nil, nil, nil, list.Table)
	return list.Container
}
func (list *EntityList) Update() {
	list.BackingStore = make([][elcNumColumns]any, 0)
	searchValid := len(list.State().SearchTerms) > 0
	searchFields := strings.Fields(list.State().SearchTerms)
	searchEntities := make([]ecs.Entity, 0)
	for _, f := range searchFields {
		if e, err := ecs.ParseEntityHumanOrCanonical(f); err == nil {
			searchEntities = append(searchEntities, e)
		}
	}
	list.State().Universe.Entities.Range(func(entity uint32) {
		if entity == 0 {
			return
		}
		rowColor := theme.Color(theme.ColorNameForeground)
		parentDesc := ""
		rank := 0
		if slices.Contains(searchEntities, ecs.Entity(entity)) {
			rank = 100
		}

		allSystem := true
		for _, c := range list.State().Universe.AllComponents(ecs.Entity(entity)) {
			if c == nil || c.Base().Flags&ecs.ComponentHideInEditor != 0 {
				continue
			}
			allSystem = false
			desc := c.String()

			if searchValid {
				rank += fuzzy.RankMatchFold(list.State().SearchTerms, desc)
			}

			if len(parentDesc) > 0 {
				parentDesc += "|"
			}
			parentDesc += desc

			switch c.(type) {
			case *core.Body:
				rowColor = colornames.Lightblue
			case *core.Sector:
				rowColor = colornames.Lightgreen
			case *materials.Image:
				rowColor = colornames.Lightpink
			case *materials.Sprite:
				rowColor = colornames.Lightpink
			case *materials.Shader:
				rowColor = colornames.Lightpink
			}
		}
		if len(parentDesc) == 0 || allSystem {
			return
		}
		dispRank := 100
		if searchValid {
			dispRank = concepts.Max(concepts.Min(rank+50, 100), 0)
		}
		backingRow := [5]any{
			int(entity),
			nil,
			parentDesc,
			dispRank,
			rowColor,
		}
		list.BackingStore = append(list.BackingStore, backingRow)
	})

	list.applySort()
}

func (list *EntityList) Select(selection *selection.Selection) {
	// TODO: Support multiple-selection when Fyne Table supports them
	for i, row := range list.BackingStore {
		for _, s := range selection.Exact {
			if row[elcEntity].(int) == int(s.Entity) {
				go fyne.Do(func() {
					// Save and restore the handlers to avoid recursive selection
					fs1 := list.Table.OnSelected
					fs2 := list.Table.OnUnselected
					list.Table.OnSelected = nil
					list.Table.OnUnselected = nil
					list.Table.Select(widget.TableCellID{Row: i, Col: 0})
					list.Table.ScrollTo(widget.TableCellID{Row: i, Col: 0})
					list.Table.OnSelected = fs1
					list.Table.OnUnselected = fs2
				})
				return
			}
		}
	}
}

func (list *EntityList) SetSort(col int, dir elSortDir) {
	for i := 0; i < int(elcNumColumns); i++ {
		list.Sorts[i] = elsdSortOff
	}
	list.Sorts[col] = dir
}

func (list *EntityList) AdvanceSort(col int) {
	order := list.Sorts[col]
	order++
	if order > elsdSortDesc {
		order = elsdSortOff
	}
	list.SetSort(col, order)
}

func (list *EntityList) applySort() {
	var order elSortDir
	var col int
	for i := range int(elcNumColumns) {
		if list.Sorts[i] != elsdSortOff {
			order = list.Sorts[i]
			col = i
			break
		}
	}

	sort.Slice(list.BackingStore, func(i, j int) bool {
		a := list.BackingStore[i]
		b := list.BackingStore[j]
		// re-sort with no sort selected
		if order == elsdSortOff {
			return a[0].(int) < b[0].(int)
		}
		switch col {
		case 0, 1, 2:
			if order == elsdSortDesc {
				return a[col].(int) > b[col].(int)
			}
			return a[col].(int) < b[col].(int)
		default:
			if order == elsdSortAsc {
				return strings.Compare(a[col].(string), b[col].(string)) < 0
			}
			return strings.Compare(a[col].(string), b[col].(string)) > 0
		}
	})
	if list.Table != nil {
		fyne.Do(func() { list.Table.Refresh() })
	}
}
