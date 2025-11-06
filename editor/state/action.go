// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import "tlyakhov/gofoom/ecs"

// Actionable represents a generic editor action.
type Actionable interface {
	ecs.Serializable
	Activate()
}

type Cancelable interface {
	Cancel()
	Status() string
}

type Action struct {
	IEditor
}

func (a *Action) Construct(data map[string]any) {}

func (a *Action) Serialize() map[string]any {
	return nil
}

func (a *Action) IsAttached() bool { return true }
