// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"log"
	"math"
	"tlyakhov/gofoom/concepts"
)

type QuadNode struct {
	Min, Max  concepts.Vector2
	MaxRadius float64 // Maximum body radius in this node
	Dead      bool

	Bodies  []*Body
	Lights  []*Body
	Sectors []*Sector

	Tree     *Quadtree
	Parent   *QuadNode
	Children [4]*QuadNode
}

func (node *QuadNode) IsLeaf() bool {
	return node.Children[0] == nil
}

func (node *QuadNode) Contains(p *concepts.Vector2) bool {
	return p[0] >= node.Min[0] && p[0] < node.Max[0] &&
		p[1] >= node.Min[1] && p[1] < node.Max[1]
}

func (node *QuadNode) Contains3D(p *concepts.Vector3) bool {
	return p[0] >= node.Min[0] && p[0] < node.Max[0] &&
		p[1] >= node.Min[1] && p[1] < node.Max[1]
}

func (node *QuadNode) recalcRadii() {
	test := node
	if test.IsLeaf() {
		test = test.Parent
	}
	for test != nil {
		test.MaxRadius = 0
		for i := range 4 {
			if test.Children[i].MaxRadius > test.MaxRadius {
				test.MaxRadius = test.Children[i].MaxRadius
			}
		}
		test = test.Parent
	}
}

func (node *QuadNode) Remove(body *Body) {
	if node.Dead {
		panic("wtf")
	}
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
		log.Printf("QuadNode.Remove: Tried to remove body %v from a node that doesn't include it.", body.Entity)
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
	parent := node.Parent
	// Are we at the root?
	if parent == nil {
		return
	}

	parent.recalcRadii()

	// Next, check if all siblings are empty to move the leaf up.

	c := &parent.Children
	sum := len(c[0].Bodies) + len(c[1].Bodies) + len(c[2].Bodies) + len(c[3].Bodies)

	if sum >= 4 {
		return
	}

	parent.Bodies = make([]*Body, sum)
	copied := 0
	for i := range 4 {
		copied += copy(parent.Bodies[copied:], c[i].Bodies)
		c[i].Bodies = nil
	}
	sumLights := len(c[0].Lights) + len(c[1].Lights) + len(c[2].Lights) + len(c[3].Lights)
	parent.Lights = make([]*Body, sumLights)
	copied = 0
	for i := range 4 {
		copied += copy(parent.Lights[copied:], c[i].Lights)
		c[i].Lights = nil
	}

	for i := 0; i < sum; i++ {
		parent.Bodies[i].QuadNode = parent
	}
	parent.Children[0] = nil
	parent.Children[1] = nil
	parent.Children[2] = nil
	parent.Children[3] = nil
	node.Dead = true
	node.Bodies = nil
	node.Lights = nil
	node.Tree = nil
}

func (node *QuadNode) increaseRadii(r float64) {
	for node != nil && r > node.MaxRadius {
		node.MaxRadius = r
		node = node.Parent
	}
}
func (node *QuadNode) subdivide() {
	halfx := (node.Min[0] + node.Max[0]) * 0.5
	halfy := (node.Min[1] + node.Max[1]) * 0.5
	node.Children[0] = &QuadNode{Min: node.Min, Parent: node, Tree: node.Tree}
	node.Children[0].Max[0] = halfx
	node.Children[0].Max[1] = halfy
	node.Children[1] = &QuadNode{Min: node.Min, Max: node.Max, Parent: node, Tree: node.Tree}
	node.Children[1].Min[0] = halfx
	node.Children[1].Max[1] = halfy
	node.Children[2] = &QuadNode{Min: node.Min, Max: node.Max, Parent: node, Tree: node.Tree}
	node.Children[2].Max[0] = halfx
	node.Children[2].Min[1] = halfy
	node.Children[3] = &QuadNode{Max: node.Max, Parent: node, Tree: node.Tree}
	node.Children[3].Min[0] = halfx
	node.Children[3].Min[1] = halfy

	log.Printf("Subdivide %v -> %v", node.Min.StringHuman(), node.Max.StringHuman())
	for i := range 4 {
		log.Printf("Child %v: %v -> %v", i, node.Children[i].Min.StringHuman(), node.Children[i].Max.StringHuman())
	}
}

func (node *QuadNode) insert(body *Body, depth int) {
	if node.Dead {
		panic("wtf")
	}

	leaf := node.IsLeaf()
	contains := node.Contains3D(&body.Pos.Now)
	switch {
	case node.Parent != nil && !contains:
		log.Printf("QuadNode.insert: body %v is not contained in this node, but also not at root?", body.Entity)
		return
	case node.Parent == nil && !contains:
		if node != node.Tree.Root {
			log.Printf("QuadNode.insert: parent is nil but this node isn't root :(")
			return
		}
		// We need to expand the tree outwards.
		newRoot := &QuadNode{
			Tree:   node.Tree,
			Parent: nil,
		}
		centerx := (node.Min[0] + node.Max[0]) * 0.5
		centery := (node.Min[1] + node.Max[1]) * 0.5
		index := 0
		if body.Pos.Now[0] < centerx {
			index |= 1
			newRoot.Min[0] = node.Min[0] - (node.Max[0] - node.Min[0])
			newRoot.Max[0] = node.Max[0]
		} else {
			newRoot.Min[0] = node.Min[0]
			newRoot.Max[0] = node.Max[0] + (node.Max[0] - node.Min[0])
		}
		if body.Pos.Now[1] < centery {
			index |= 2
			newRoot.Min[1] = node.Min[1] - (node.Max[1] - node.Min[1])
			newRoot.Max[1] = node.Max[1]
		} else {
			newRoot.Min[1] = node.Min[1]
			newRoot.Max[1] = node.Max[1] + (node.Max[1] - node.Min[1])
		}
		newRoot.subdivide()
		newRoot.Children[index] = node
		node.Parent = newRoot
		node.Tree.Root = newRoot
		newRoot.insert(body, 0)
		return
	case leaf && (len(node.Bodies) < 4 || depth > 8):
		if !contains {
			log.Printf("QuadNode.insert: reached leaf, but node doesn't contain body.")
			return
		}
		// We can insert here!
		if body.Pos.Now[2] < node.Tree.MinZ {
			node.Tree.MinZ = body.Pos.Now[2]
		}
		if body.Pos.Now[2] > node.Tree.MaxZ {
			node.Tree.MaxZ = body.Pos.Now[2]
		}
		node.increaseRadii(body.Size.Now[0] * 0.5)
		node.Bodies = append(node.Bodies, body)
		if light := GetLight(body.ECS, body.Entity); light != nil {
			node.Lights = append(node.Lights, body)
		}
		body.QuadNode = node
		return
	case leaf:
		// Full, need to break it up
		node.subdivide()
		oldBodies := node.Bodies
		node.Bodies = nil
		node.Lights = nil
		// We are no longer a leaf. Insert into children
		for _, b := range oldBodies {
			b.QuadNode = nil
			for _, child := range node.Children {
				if child.Contains3D(&body.Pos.Now) {
					child.insert(body, depth+1)
					return
				}
			}
			if b.QuadNode == nil {
				node.Children[0].Bodies = append(node.Children[0].Bodies, b)
			}
		}
	}

	for _, child := range node.Children {
		if child.Contains3D(&body.Pos.Now) {
			child.insert(body, depth+1)
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
	dz := node.Tree.MinZ - center[2]
	dy := node.Min[1] - center[1]
	dx := node.Min[0] - center[0]
	dz2 := node.Tree.MaxZ - center[2]
	dy2 := node.Max[1] - center[1]
	dx2 := node.Max[0] - center[0]
	return dx*normal[0]+dy*normal[1]+dz*normal[2] >= 0 ||
		dx2*normal[0]+dy2*normal[1]+dz2*normal[2] >= 0
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

func (node *QuadNode) rayOverlaps(start, dir *concepts.Vector3) bool {
	invx := 1.0 / dir[0]
	invy := 1.0 / dir[1]

	tx1 := (node.Min[0] - start[0]) * invx
	tx2 := (node.Max[0] - start[0]) * invx

	tmin := math.Min(tx1, tx2)
	tmax := math.Max(tx1, tx2)

	ty1 := (node.Min[1] - start[1]) * invy
	ty2 := (node.Max[1] - start[1]) * invy

	tmin = math.Max(tmin, math.Min(ty1, ty2))
	tmax = math.Min(tmax, math.Max(ty1, ty2))

	return tmax >= math.Max(tmin, 0.0)
}

func (node *QuadNode) RangeRay(start, dir *concepts.Vector3, lightsOnly bool, fn func(b *Body) bool) {
	if node.IsLeaf() {
		slice := node.Bodies
		if lightsOnly {
			slice = node.Lights
		}
		for _, b := range slice {
			if !fn(b) {
				break
			}
		}
		return
	}
	for i := range 4 {
		if !node.Children[i].rayOverlaps(start, dir) {
			continue
		}
		node.Children[i].RangeRay(start, dir, lightsOnly, fn)
	}
}
