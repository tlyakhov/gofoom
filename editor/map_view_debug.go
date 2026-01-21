// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/pathfinding"
)

func (mw *MapWidget) drawPath(path []concepts.Vector3) {
	if len(path) <= 1 {
		return
	}
	mw.Context.SetRGBA(0.5, 0.5, 1.0, 1.0)
	mw.Context.NewSubPath()
	mw.Context.MoveTo(path[0][0], path[0][1])
	for i, v := range path {
		if i == 0 {
			continue
		}
		mw.Context.LineTo(v[0], v[1])
	}
	mw.Context.Stroke()
}

func (mw *MapWidget) drawPaths() {
	if editor.PathDebugStart != editor.PathDebugEnd {
		pf := pathfinding.Finder{Start: &editor.PathDebugStart,
			End:    &editor.PathDebugEnd,
			Radius: 16,
			Step:   10}
		mw.drawPath(pf.ShortestPath())
	}

	arena := ecs.ArenaFor[behaviors.ActorState](behaviors.ActorStateCID)
	for i := range arena.Cap() {
		state := arena.Value(i)
		if state == nil || len(state.Path) == 0 {
			continue
		}
		mw.drawPath(state.Path)
	}

}

func (mw *MapWidget) drawPursuers() {
	arena := ecs.ArenaFor[behaviors.Pursuer](behaviors.PursuerCID)
	for i := range arena.Cap() {
		pursuer := arena.Value(i)
		if pursuer == nil {
			continue
		}
		mw.Context.SetRGBA(0.5, 0.5, 1.0, 1.0)
		/*for _, e := range pursuer.Breadcrumbs {
			if e.Key == 0 {
				continue
			}
			mw.DrawHandle(e.Data.Pos.To2D())
		}*/

		//best := pursuer.BestCandidate()

		b := core.GetBody(pursuer.Entity)
		for i, c := range pursuer.Candidates {
			if c == nil || c.Count == 0 {
				continue
			}
			w := c.Weight // / float64(c.Count)
			ray := &concepts.Ray{Start: b.Pos.Render}
			ray.FromAngleAndLimit(float64(i)*360/float64(len(pursuer.Candidates)), 0, math.Abs(w)*64)
			opacity := 1.0
			//if best == c {
			//	opacity = 1
			//}
			if w > 0 {
				mw.Context.SetRGBA(0, 1, 0, opacity)
			} else {
				mw.Context.SetRGBA(1, 0, 0, opacity)
			}
			mw.Context.NewSubPath()
			mw.Context.MoveTo(ray.Start[0], ray.Start[1])
			mw.Context.LineTo(ray.End[0], ray.End[1])
			mw.Context.Stroke()
		}
	}
}
