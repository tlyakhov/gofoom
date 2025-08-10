// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"log"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

type Quadtree struct {
	ecs.Attached `editable:"^"`

	MinZ, MaxZ float64

	Root *QuadNode
}

func (q *Quadtree) String() string {
	return "Quadtree"
}

func (q *Quadtree) Construct(data map[string]any) {
	q.Attached.Construct(data)

	q.Flags |= ecs.ComponentInternal // never serialize this
	q.Build()
}

func (q *Quadtree) Serialize() map[string]any {
	result := q.Attached.Serialize()

	return result
}

func (q *Quadtree) Update(body *Body) {
	if body.QuadNode == nil {
		q.Root.insert(body, 0)
		return
	} else if body.QuadNode.Contains3D(&body.Pos.Now) {
		return
	}

	body.QuadNode.Remove(body)
	body.QuadNode = nil
	q.Root.insert(body, 0)
}

func (q *Quadtree) Build() {
	q.Root = &QuadNode{Tree: q}

	offset := constants.QuadtreeInitDim / 16
	q.Root.Min[0] = offset
	q.Root.Min[1] = offset
	q.MinZ = 0
	q.Root.Max[0] = offset + constants.QuadtreeInitDim
	q.Root.Max[1] = offset + constants.QuadtreeInitDim
	q.MaxZ = 0

	colBody := ecs.ArenaFor[Body](BodyCID)
	for i := range colBody.Cap() {
		body := colBody.Value(i)
		if body == nil || !body.IsActive() {
			continue
		}
		q.Update(body)
	}
}

func (q *Quadtree) Print() {
	log.Println("Tree:")
	q.Root.print(0)
}
