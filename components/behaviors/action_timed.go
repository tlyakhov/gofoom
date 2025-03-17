// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type ActionTimed struct {
	ecs.Attached `editable:"^"`

	Delay dynamic.DynamicValue[float64] `editable:"Delay"`

	// TODO: Add ability to repeat multiple times in a single action
	// TODO: Add ability to fire at random intervals
	Fired containers.Set[ecs.Entity]
}

func (timed *ActionTimed) String() string {
	return "Timed"
}

func (timed *ActionTimed) OnDelete() {
	if timed.Universe != nil {
		timed.Delay.Detach(timed.Universe.Simulation)
	}
	timed.Attached.OnDelete()
}

func (timed *ActionTimed) OnAttach(u *ecs.Universe) {
	timed.Attached.OnAttach(u)
	timed.Delay.Attach(u.Simulation)
}

func (timed *ActionTimed) Construct(data map[string]any) {
	timed.Attached.Construct(data)

	timed.Delay.SetAll(0)
	timed.Fired = make(containers.Set[ecs.Entity])

	if data == nil {
		return
	}

	if v, ok := data["Delay"]; ok {
		timed.Delay.Construct(v.(map[string]any))
	}
}

func (timed *ActionTimed) Serialize() map[string]any {
	result := timed.Attached.Serialize()

	if timed.Delay.Spawn != 0 {
		result["Delay"] = timed.Delay.Serialize()
	}

	return result
}
