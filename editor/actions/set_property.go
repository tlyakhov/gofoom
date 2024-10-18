// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"fmt"
	"log"
	"reflect"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type SetProperty struct {
	state.IEditor
	*state.PropertyGridField
	Original       []reflect.Value
	ValuesToAssign []reflect.Value
}

func (a *SetProperty) AssignAll(v reflect.Value) {
	a.ValuesToAssign = make([]reflect.Value, len(a.Values))
	for i := range a.ValuesToAssign {
		a.ValuesToAssign[i] = v
	}
}

func (a *SetProperty) FireHooks() {
	// TODO: Optimize this by remembering visited parents to avoid firing these
	// multiple times for the same selection.
	for _, v := range a.Values {
		switch target := v.Parent().(type) {
		case dynamic.Dynamic:
			target.ResetToOriginal()
			target.Recalculate()
		case *materials.Image:
			if a.Source.Name == "Source" {
				target.Load()
			}
		case *core.Script:
			// TODO: use a nicer source code editor for script properties.
			target.Compile()
		case *materials.Text:
			target.RasterizeText()
		case *ecs.Instanced:
			target.Recalculate()
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
			} else if a.Name == "Segment.Portal sector" {
				target.Recalculate()
			}

			log.Printf("%v", a.Name)
		}
	}
}

func (a *SetProperty) Act() {
	defer a.ActionFinished(false, true, false)
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for i, v := range a.Values {
		origValue := reflect.ValueOf(v.Interface())
		a.Original = append(a.Original, origValue)
		v.Deref().Set(a.ValuesToAssign[i])
	}

	a.FireHooks()
	a.State().Modified = true
}

func (a *SetProperty) Undo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for i, v := range a.Values {
		fmt.Printf("Undo: %v\n", a.Original[i].String())
		v.Deref().Set(a.Original[i])
	}
	a.FireHooks()
}
func (a *SetProperty) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for i, v := range a.Values {
		fmt.Printf("Redo: %v\n", a.ValuesToAssign[i].String())
		v.Deref().Set(a.ValuesToAssign[i])
	}
	a.FireHooks()
}
