// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"math"
	"tlyakhov/gofoom/ecs"
)

type Quadtree struct {
	ecs.Attached `editable:"^"`

	Root *QuadNode
}

var QuadtreeCID ecs.ComponentID

func init() {
	QuadtreeCID = ecs.RegisterComponent(&ecs.Column[Quadtree, *Quadtree]{Getter: GetQuadtree})
}

func GetQuadtree(db *ecs.ECS, e ecs.Entity) *Quadtree {
	if asserted, ok := db.Component(e, QuadtreeCID).(*Quadtree); ok {
		return asserted
	}
	return nil
}

func (q *Quadtree) String() string {
	return "Quadtree"
}

func (q *Quadtree) Construct(data map[string]any) {
	q.Attached.Construct(data)
}

func (q *Quadtree) Serialize() map[string]any {
	result := q.Attached.Serialize()

	return result
}

func (q *Quadtree) Build() {
	q.Root = &QuadNode{}

	// Find overall min/max
	q.Root.Min[0] = math.Inf(1)
	q.Root.Min[1] = math.Inf(1)
	q.Root.Max[0] = math.Inf(-1)
	q.Root.Max[1] = math.Inf(-1)
	col := ecs.ColumnFor[Sector](q.ECS, SectorCID)
	for i := range col.Cap() {
		sector := col.Value(i)
		if sector == nil {
			continue
		}
		if sector.Min[0] < q.Root.Min[0] {
			q.Root.Min[0] = sector.Min[0]
		}
		if sector.Max[0] > q.Root.Max[0] {
			q.Root.Max[0] = sector.Max[0]
		}
		if sector.Min[1] < q.Root.Min[1] {
			q.Root.Min[1] = sector.Min[1]
		}
		if sector.Max[1] > q.Root.Max[1] {
			q.Root.Max[1] = sector.Max[1]
		}
	}

	colBody := ecs.ColumnFor[Body](q.ECS, BodyCID)
	for i := range colBody.Cap() {
		body := colBody.Value(i)
		if body == nil {
			continue
		}
		q.Root.Insert(body)
	}
}

func TheQuadtree(db *ecs.ECS) *Quadtree {
	a := db.First(QuadtreeCID)
	if a != nil {
		return a.(*Quadtree)
	}

	q := db.NewAttachedComponent(db.NewEntity(), QuadtreeCID).(*Quadtree)
	q.System = true
	q.Build()
	return q
}
