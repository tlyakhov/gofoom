// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"fmt"
	"reflect"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/driver/desktop"
)

type SetProperty struct {
	state.IEditor
	*state.PropertyGridField
	Original []reflect.Value
	ToSet    reflect.Value
}

func (a *SetProperty) OnMouseDown(evt *desktop.MouseEvent) {}
func (a *SetProperty) OnMouseMove()                        {}
func (a *SetProperty) OnMouseUp()                          {}
func (a *SetProperty) Cancel()                             {}
func (a *SetProperty) Frame()                              {}

func (a *SetProperty) FireHooks() {
	// TODO: this has a bug - all these fire for the same parent. We need to
	// be able to handle multiple chains of properties
	for range a.Values {
		switch target := a.Parent.(type) {
		case concepts.Simulated:
			target.Reset()
		case *materials.Image:
			if a.Source.Name == "Source" {
				target.Load()
			}
		case *core.Script:
			// TODO: use a nicer source code editor for script properties.
			target.Compile()
		case *materials.Text:
			target.RasterizeText()
		case *core.SectorSegment:
			// For SectorSegments, the A & B fields of the child Segment type
			// are pointers to SectorSegment.P and SectorSegment.Next.P
			// respectively. If the user edits the A or B field, we need to
			// propagate that setting.
			if a.Name == "Segment.A" {
				target.P = *target.A
				target.Recalculate()
			} else if a.Name == "Segment.B" {
				target.Next.P = *target.B
				target.Recalculate()
			}
		}
	}
}

func (a *SetProperty) Act() {
	defer a.ActionFinished(false, true, false)
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for _, v := range a.Values {
		origValue := reflect.ValueOf(v.Elem().Interface())
		a.Original = append(a.Original, origValue)
		v.Elem().Set(a.ToSet)
	}
	a.FireHooks()
	a.State().Modified = true
}

func (a *SetProperty) Undo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for i, v := range a.Values {
		fmt.Printf("Undo: %v\n", a.Original[i].String())
		v.Elem().Set(a.Original[i])
	}
	a.FireHooks()
}
func (a *SetProperty) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for _, v := range a.Values {
		fmt.Printf("Redo: %v\n", a.ToSet.String())
		v.Elem().Set(a.ToSet)
	}
	a.FireHooks()
}
