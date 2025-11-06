// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"fmt"
	"log"
	"reflect"

	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type SetProperty struct {
	state.Action
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
		if v.Entity.IsExternal() {
			continue
		}
		switch target := v.Parent().(type) {
		case *materials.Image:
			target.MarkDirty()
			ecs.ActAllControllersOneEntity(v.Entity, ecs.ControllerRecalculate)
			a.FlushEntityImage(v.Entity)
		case *ecs.Linked, *audio.Sound, *core.Script, *core.SectorPlane, *core.Sector:
			ecs.ActAllControllersOneEntity(v.Entity, ecs.ControllerRecalculate)
			a.FlushEntityImage(v.Entity)
			// TODO: use a nicer source code editor for script properties.
		case *ecs.SourceFile:
			if target.Loaded {
				target.Unload()
			}
			target.Load()
		case dynamic.Dynamic:
			target.ResetToSpawn()
			target.Recalculate()
		case *materials.Text:
			target.RasterizeText()
		case *core.SectorSegment:
			// For SectorSegments, the A & B fields of the child Segment type
			// are pointers to SectorSegment.P and SectorSegment.Next.P
			// respectively. If the user edits the A or B field, we need to
			// propagate that setting.
			switch a.Name {
			case "Segment.A":
				target.P.SetAll(*target.A)
				target.Recalculate()
			case "Segment.B":
				target.Next.P.SetAll(*target.B)
				target.Recalculate()
			case "Segment.Portal sector":
				target.Recalculate()
			}

			log.Printf("SetProperty.FireHooks for *core.SectorSegment: %v", a.Name)
		}
	}
}

func (a *SetProperty) Activate() {
	for i, v := range a.Values {
		if v.Entity.IsExternal() {
			continue
		}
		origValue := reflect.ValueOf(v.Interface())
		a.Original = append(a.Original, origValue)
		v.Deref().Set(a.ValuesToAssign[i])
	}

	a.FireHooks()
	a.State().Modified = true
	a.ActionFinished(false, true, false)
}

func (a *SetProperty) Undo() {
	for i, v := range a.Values {
		if v.Entity.IsExternal() {
			continue
		}
		fmt.Printf("Undo: %v\n", a.Original[i].String())
		v.Deref().Set(a.Original[i])
	}
	a.FireHooks()
}
func (a *SetProperty) Redo() {
	for i, v := range a.Values {
		if v.Entity.IsExternal() {
			continue
		}
		fmt.Printf("Redo: %v\n", a.ValuesToAssign[i].String())
		v.Deref().Set(a.ValuesToAssign[i])
	}
	a.FireHooks()
}
