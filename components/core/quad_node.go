// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/concepts"
)

type QuadNode struct {
	Min, Max  concepts.Vector2
	Depth     int
	MaxRadius float64 // Maximum body radius in this node

	Bodies  []*Body
	Lights  []*Body
	Sectors []*Sector

	Parent   *QuadNode
	Children [4]*QuadNode
}

func (node *QuadNode) IsLeaf() bool {
	return node.Children[0] == nil
}

func (node *QuadNode) Contains(b *Body) bool {
	return b.Pos.Now[0] >= node.Min[0] && b.Pos.Now[0] < node.Max[0] &&
		b.Pos.Now[1] >= node.Min[1] && b.Pos.Now[1] < node.Max[1]
}

func (node *QuadNode) Update(body *Body) {
	if body.QuadNode == nil {
		node.Insert(body)
		return
	} else if body.QuadNode.Contains(body) {
		return
	}

	body.QuadNode.Remove(body)
	node.Insert(body)
}

func (node *QuadNode) recalcRadii() {
	for node != nil {
		if node.IsLeaf() {
			node = node.Parent
			continue
		}
		node.MaxRadius = node.Children[0].MaxRadius

		if node.Children[1].MaxRadius > node.MaxRadius {
			node.MaxRadius = node.Children[1].MaxRadius
		}
		if node.Children[2].MaxRadius > node.MaxRadius {
			node.MaxRadius = node.Children[2].MaxRadius
		}
		if node.Children[3].MaxRadius > node.MaxRadius {
			node.MaxRadius = node.Children[3].MaxRadius
		}
		node = node.Parent
	}
}

func (node *QuadNode) Remove(body *Body) {
	found := false
	node.MaxRadius = 0
	// Find the body in the slice and trim
	for i, b := range node.Bodies {
		if b != body {
			r := b.Size.Now[0] * 0.5
			if r > node.MaxRadius {
				node.MaxRadius = r
			}
			continue
		}
		l := len(node.Bodies) - 1
		node.Bodies[i] = node.Bodies[l]
		node.Bodies = node.Bodies[:l]
		found = true
		break
	}
	if !found {
		return
	}

	if light := GetLight(body.ECS, body.Entity); light != nil {
		// Find the light in the slice and trim
		for i, b := range node.Lights {
			if b != body {
				continue
			}
			l := len(node.Lights) - 1
			node.Lights[i] = node.Lights[l]
			node.Lights = node.Lights[:l]
			break
		}
	}
	node.Parent.recalcRadii()

	// Next, check if all siblings are empty to move the leaf up.

	// Are we at the root?
	/*if node.Parent == nil {
		return
	}
	c := node.Parent.Children
	if len(c[0].Bodies) != 0 || len(c[1].Bodies) != 0 ||
		len(c[2].Bodies) != 0 || len(c[3].Bodies) != 0 {
		return
	}
	node.Parent.Children[0] = nil
	node.Parent.Children[1] = nil
	node.Parent.Children[2] = nil
	node.Parent.Children[3] = nil*/
}

func (node *QuadNode) available() bool {
	return len(node.Bodies) < 4 || node.Depth >= 8
}

func (node *QuadNode) increaseRadii(r float64) {
	for node != nil && r > node.MaxRadius {
		node.MaxRadius = r
		node = node.Parent
	}
}

func (node *QuadNode) Insert(body *Body) {
	leaf := node.IsLeaf()
	switch {
	case leaf && node.available():
		node.increaseRadii(body.Size.Now[0] * 0.5)
		node.Bodies = append(node.Bodies, body)
		if light := GetLight(body.ECS, body.Entity); light != nil {
			node.Lights = append(node.Lights, body)
		}
		body.QuadNode = node
		return
	case leaf:
		// Full, need to break it up
		halfx := node.Min[0] + (node.Max[0]-node.Min[0])/2
		halfy := node.Min[1] + (node.Max[1]-node.Min[1])/2
		node.Children[0] = &QuadNode{Min: node.Min, Max: node.Max, Parent: node, Depth: node.Depth + 1}
		node.Children[0].Max[0] = halfx
		node.Children[0].Max[1] = halfy
		node.Children[1] = &QuadNode{Min: node.Min, Max: node.Max, Parent: node, Depth: node.Depth + 1}
		node.Children[1].Min[0] = halfx
		node.Children[1].Max[1] = halfy
		node.Children[2] = &QuadNode{Min: node.Min, Max: node.Max, Parent: node, Depth: node.Depth + 1}
		node.Children[2].Min[0] = halfx
		node.Children[2].Min[1] = halfy
		node.Children[3] = &QuadNode{Min: node.Min, Max: node.Max, Parent: node, Depth: node.Depth + 1}
		node.Children[3].Max[0] = halfx
		node.Children[3].Min[1] = halfy
		oldBodies := node.Bodies
		node.Bodies = nil
		node.Lights = nil
		// We are no longer a leaf. Insert into children
		for _, b := range oldBodies {
			node.Insert(b)
		}
		node.Insert(body)
		return
	}

	for _, child := range node.Children {
		if child.Contains(body) {
			child.Insert(body)
			return
		}
	}
}

func (node *QuadNode) circleOverlaps(center *concepts.Vector2, r float64) bool {
	r += node.MaxRadius
	if center[0]+r < node.Min[0] || center[1]+r < node.Min[1] ||
		center[0]-r >= node.Max[0] || center[1]-r >= node.Max[1] {
		return false
	}
	return true
}

func (node *QuadNode) RangeCircle(center *concepts.Vector2, r float64, fn func(b *Body) bool) {
	if node.IsLeaf() {
		for _, b := range node.Bodies {
			dx := b.Pos.Now[0] - center[0]
			dy := b.Pos.Now[1] - center[1]
			br := b.Size.Now[0] * 0.5
			if dx*dx+dy*dy > (r+br)*(r+br) {
				continue
			}
			if !fn(b) {
				break
			}
		}
		return
	}
	for i := range 4 {
		if !node.Children[i].circleOverlaps(center, r) {
			continue
		}
		node.Children[i].RangeCircle(center, r, fn)
	}
}

func (node *QuadNode) planeOverlaps(center, normal *concepts.Vector3) bool {
	if normal[0] == 0 && normal[1] == 0 {
		return true
	}
	dx := node.Min[0] - center[0]
	dy := node.Min[1] - center[1]
	dx2 := node.Max[0] - center[0]
	dy2 := node.Max[1] - center[1]
	return dx*normal[0]+dy*normal[1] >= 0 ||
		dx2*normal[0]+dy2*normal[1] >= 0
}

func (node *QuadNode) RangePlane(center, normal *concepts.Vector3, lightsOnly bool, fn func(b *Body) bool) {
	if node.IsLeaf() {
		slice := node.Bodies
		if lightsOnly {
			slice = node.Lights
		}
		for _, b := range slice {
			dx := b.Pos.Now[0] - center[0]
			dy := b.Pos.Now[1] - center[1]
			dz := b.Pos.Now[2] - center[2]
			if dx*normal[0]+dy*normal[1]+dz*normal[2] < 0 {
				continue
			}
			if !fn(b) {
				break
			}
		}
		return
	}
	for i := range 4 {
		if !node.Children[i].planeOverlaps(center, normal) {
			continue
		}
		node.Children[i].RangePlane(center, normal, lightsOnly, fn)
	}
}
