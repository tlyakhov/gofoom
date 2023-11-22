package properties

import (
	"log"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type EntityTreeColumnID int

//go:generate go run github.com/dmarkham/enumer -type=EntityTreeColumnID -json
const (
	etcEntity EntityTreeColumnID = iota
	etcIndex
	etcDesc
	etcColor
	etcRank
)

type EntityTree struct {
	state.IEditor
	View      *gtk.TreeView
	Store     *gtk.TreeStore
	SortModel *gtk.TreeModelSort
	Columns   []*gtk.TreeViewColumn
}

func (et *EntityTree) Filter() {
	et.UpdateSearch()
}

func (et *EntityTree) SelectionChanged(sel *gtk.TreeSelection) {
	list := sel.GetSelectedRows(et.SortModel)
	iter := list.First()
	ers := make([]any, 0)
	for iter != nil {
		treePath := iter.Data().(*gtk.TreePath)
		iter2, _ := et.SortModel.GetIter(treePath)
		v, _ := et.SortModel.GetValue(iter2, int(etcEntity))
		gv, _ := v.GoValue()
		ers = append(ers, et.State().DB.EntityRef(gv.(uint64)))
		iter = iter.Next()
	}
	et.IEditor.SelectObjects(ers)
}

func (et *EntityTree) ColumnClicked(col *gtk.TreeViewColumn) {
	curId, _, _ := et.SortModel.GetSortColumnId()
	clickedId := col.GetSortColumnID()
	order := col.GetSortOrder()
	log.Printf("current: %v, %v. clicked: %v", curId, order, clickedId)

	if clickedId == curId && order != gtk.SORT_ASCENDING {
		et.SortModel.SetSortColumnId(clickedId, gtk.SORT_DESCENDING)
		col.SetSortIndicator(true)
		//col.SetSortOrder(gtk.SORT_DESCENDING)
		return
	} else if clickedId == curId && order != gtk.SORT_DESCENDING {
		et.SortModel.SetSortColumnId(clickedId, gtk.SORT_ASCENDING)
		col.SetSortIndicator(true)
		//col.SetSortOrder(gtk.SORT_ASCENDING)
		return
	} else {
		if curId >= 0 && curId < len(et.Columns) {
			et.Columns[curId].SetSortIndicator(false)
		}
		et.SortModel.SetSortColumnId(clickedId, gtk.SORT_ASCENDING)
		col.SetSortIndicator(true)
		//col.SetSortOrder(gtk.SORT_ASCENDING)
		return
	}
}
func (et *EntityTree) addColumn(col int, rend gtk.ICellRenderer) {
	et.Columns[col].AddAttribute(rend, "foreground", int(etcColor))
	et.Columns[col].SetClickable(true)
	if col == 1 {
		et.Columns[col].SetExpand(true)
		et.Columns[col].SetSizing(gtk.TREE_VIEW_COLUMN_AUTOSIZE)
	}
	et.Columns[col].Connect("clicked", et.ColumnClicked)
	et.View.AppendColumn(et.Columns[col])

}
func (et *EntityTree) Construct() {
	et.Store, _ = gtk.TreeStoreNew(glib.TYPE_UINT64, glib.TYPE_INT, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_INT)
	et.SortModel, _ = gtk.TreeModelSortNew(et.Store)
	et.View.SetModel(et.SortModel)
	sel, _ := et.View.GetSelection()
	sel.SetMode(gtk.SELECTION_MULTIPLE)
	sel.Connect("changed", et.SelectionChanged)

	et.SortModel.SetSortFunc(int(etcEntity), func(model *gtk.TreeModel, a, b *gtk.TreeIter) int {
		return int(et.nodeValue(a, etcEntity).(uint64) - et.nodeValue(b, etcEntity).(uint64))
	})

	et.SortModel.SetSortFunc(int(etcIndex), func(model *gtk.TreeModel, a, b *gtk.TreeIter) int {
		return et.nodeValue(a, etcIndex).(int) - et.nodeValue(b, etcIndex).(int)
	})

	et.SortModel.SetSortFunc(int(etcDesc), func(model *gtk.TreeModel, a, b *gtk.TreeIter) int {
		sa := et.nodeValue(a, etcDesc).(string)
		sb := et.nodeValue(b, etcDesc).(string)
		if sa < sb {
			return -1
		} else if sb > sa {
			return 1
		}
		return 0
	})

	et.SortModel.SetSortFunc(int(etcRank), func(model *gtk.TreeModel, a, b *gtk.TreeIter) int {
		return et.nodeValue(a, etcRank).(int) - et.nodeValue(b, etcRank).(int)
	})

	et.Columns = make([]*gtk.TreeViewColumn, 3)
	rend, _ := gtk.CellRendererTextNew()
	et.Columns[0], _ = gtk.TreeViewColumnNewWithAttribute("Entity", rend, "text", int(etcEntity))
	et.Columns[0].SetSortColumnID(int(etcEntity))
	et.addColumn(0, rend)

	rend2, _ := gtk.CellRendererTextNew()
	et.Columns[1], _ = gtk.TreeViewColumnNewWithAttribute("Description", rend2, "text", int(etcDesc))
	et.Columns[1].SetSortColumnID(int(etcDesc))
	et.addColumn(1, rend2)

	rend3, _ := gtk.CellRendererProgressNew()
	et.Columns[2], _ = gtk.TreeViewColumnNewWithAttribute("Search Rank", rend3, "value", int(etcRank))
	et.Columns[2].SetSortColumnID(int(etcRank))
	et.addColumn(2, rend3)
	et.Columns[2].SetMaxWidth(90)
	et.Columns[2].SetExpand(false)
	et.Columns[2].SetSizing(gtk.TREE_VIEW_COLUMN_FIXED)

	et.ColumnClicked(et.Columns[2])
	et.ColumnClicked(et.Columns[2])
}

func (et *EntityTree) nodeValue(iter *gtk.TreeIter, col EntityTreeColumnID) interface{} {
	v, _ := et.Store.GetValue(iter, int(col))
	gv, _ := v.GoValue()
	return gv
}

func (et *EntityTree) updateSearchChild(iter *gtk.TreeIter) int {
	valid := true
	totalRank := 0
	for valid {
		desc := et.nodeValue(iter, etcDesc).(string)
		rank := 0
		if len(et.State().Filter) > 0 {
			rank = fuzzy.RankMatchFold(et.State().Filter, desc)
		}
		n := et.Store.IterNChildren(iter)
		if n > 0 {
			child, _ := et.Store.GetIterFirst()
			et.Store.IterChildren(iter, child)
			rank += et.updateSearchChild(child)
		}

		dispRank := concepts.Max(concepts.Min(rank+50, 100), 0)
		et.Store.SetValue(iter, int(etcRank), dispRank)
		totalRank += rank
		valid = et.Store.IterNext(iter)
	}
	return totalRank
}
func (et *EntityTree) UpdateSearch() {
	iter, valid := et.Store.GetIterFirst()
	if !valid {
		return
	}
	et.updateSearchChild(iter)
}

func (et *EntityTree) Update() {
	et.Store.Clear()
	index := 0
	for entity, components := range et.State().DB.EntityComponents {
		iter := et.Store.Append(nil)
		et.Store.SetValue(iter, int(etcEntity), entity)
		parentDesc := ""

		for index, c := range components {
			if c == nil {
				continue
			}
			child := et.Store.Append(iter)
			desc := c.String()
			et.Store.SetValue(child, int(etcEntity), entity)
			et.Store.SetValue(child, int(etcIndex), index)
			et.Store.SetValue(child, int(etcDesc), desc)

			if len(parentDesc) > 0 {
				parentDesc += "|"
			}
			parentDesc += desc

			if index == core.BodyComponentIndex {
				et.Store.SetValue(iter, int(etcColor), "lightblue")
				et.Store.SetValue(child, int(etcColor), "lightblue")
			} else if index == core.SectorComponentIndex {
				et.Store.SetValue(iter, int(etcColor), "lightgreen")
				et.Store.SetValue(child, int(etcColor), "lightgreen")
			} else if index == materials.ImageComponentIndex || index == materials.SolidComponentIndex {
				et.Store.SetValue(iter, int(etcColor), "lightpink")
				et.Store.SetValue(child, int(etcColor), "lightpink")
			}

		}
		et.Store.SetValue(iter, int(etcDesc), parentDesc)
		index++
	}
	et.UpdateSearch()
}
