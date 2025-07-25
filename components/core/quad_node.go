// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"log"
	"math"
	"strings"
	"tlyakhov/gofoom/concepts"
)

type QuadNode struct {
	Min, Max  concepts.Vector2
	MaxRadius float64 // Maximum body radius in this node
	Bodies    []*Body
	Lights    []*Body

	Tree     *Quadtree
	Parent   *QuadNode
	Children [4]*QuadNode
}

func (node *QuadNode) print(depth int) {
	ds := strings.Repeat("  ", depth)
	if node.IsLeaf() {
		log.Printf("%vLeaf %p: %v -> %v (%v bodies, %v lights)",
			ds, node,
			node.Min.StringHuman(), node.Max.StringHuman(),
			node.BodyEntitiesString(), len(node.Lights))
	} else {
		log.Printf("%vNode %p: %v -> %v", ds, node, node.Min.StringHuman(), node.Max.StringHuman())
		for i := range 4 {
			node.Children[i].print(depth + 1)
		}
	}
}

func (node *QuadNode) BodyEntitiesString() string {
	result := ""
	for _, b := range node.Bodies {
		if result != "" {
			result += ";"
		}
		result += b.Entities.String()
	}
	return result
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
	if test != nil && test.Children[0] == nil {
		log.Printf("QuadNode.recalcRadii: multiple leaves in a row?")
		return
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
	if !node.IsLeaf() {
		log.Printf("QuadNode.Remove: node is not a leaf!")
	}

	foundIndex := -1
	node.MaxRadius = 0
	// Find the body in the slice and trim
	for i, test := range node.Bodies {
		if test == body {
			// Don't break, we still need to keep track of radii
			foundIndex = i
			continue
		}
		r := test.Size.Now[0] * 0.5
		if r > node.MaxRadius {
			node.MaxRadius = r
		}
	}
	if foundIndex < 0 {
		log.Printf("QuadNode.Remove: Tried to remove body %v from a node that doesn't include it.", body.Entity)
		return
	}
	size := len(node.Bodies) - 1
	node.Bodies[foundIndex] = node.Bodies[size]
	node.Bodies = node.Bodies[:size]
	body.QuadNode = nil

	if light := GetLight(body.Entity); light != nil {
		// Find the light in the slice and trim
		for i, test := range node.Lights {
			if test != body {
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

	if c[0] != node && c[1] != node && c[2] != node && c[3] != node {
		log.Printf("QuadNode.Remove: node is not part of parent's children.")
		return
	}

	sum := 0

	for i := range 4 {
		if !c[i].IsLeaf() {
			// Can't remove children because not all are leaves.
			return
		}
		sum += len(c[i].Bodies)
	}

	if sum >= 4 {
		return
	}

	if len(parent.Bodies) > 0 {
		log.Printf("QuadNode.Remove: non-leaf has bodies?")
	}

	parent.Bodies = make([]*Body, 0, sum)
	for i := range 4 {
		if len(c[i].Bodies) == 0 {
			continue
		}
		parent.Bodies = append(parent.Bodies, c[i].Bodies...)
		c[i].Bodies = nil
	}
	sumLights := len(c[0].Lights) + len(c[1].Lights) + len(c[2].Lights) + len(c[3].Lights)
	parent.Lights = make([]*Body, 0, sumLights)
	for i := range 4 {
		if len(c[i].Lights) == 0 {
			continue
		}
		parent.Lights = append(parent.Lights, c[i].Lights...)
		c[i].Lights = nil
	}

	for i := 0; i < sum; i++ {
		parent.Bodies[i].QuadNode = parent
	}
	parent.Children[0] = nil
	parent.Children[1] = nil
	parent.Children[2] = nil
	parent.Children[3] = nil
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
}

func (node *QuadNode) expandRoot(pos *concepts.Vector3) {
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
	if pos[0] < centerx {
		index |= 1
		newRoot.Min[0] = node.Min[0] - (node.Max[0] - node.Min[0])
		newRoot.Max[0] = node.Max[0]
	} else {
		newRoot.Min[0] = node.Min[0]
		newRoot.Max[0] = node.Max[0] + (node.Max[0] - node.Min[0])
	}
	if pos[1] < centery {
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
}

func (node *QuadNode) addToLeaf(body *Body) {
	if body.Pos.Now[2] < node.Tree.MinZ {
		node.Tree.MinZ = body.Pos.Now[2] - body.Size.Now[1]
	}
	if body.Pos.Now[2] > node.Tree.MaxZ {
		node.Tree.MaxZ = body.Pos.Now[2] + body.Size.Now[1]
	}
	node.increaseRadii(body.Size.Now[0] * 0.5)
	node.Bodies = append(node.Bodies, body)
	if light := GetLight(body.Entity); light != nil {
		node.Lights = append(node.Lights, body)
	}
	body.QuadNode = node
}

func (node *QuadNode) increaseDepth(depth int) {
	node.subdivide()
	oldBodies := node.Bodies
	node.Bodies = nil
	node.Lights = nil
	// We are no longer a leaf. Insert into children
	for _, oldBody := range oldBodies {
		found := false
		for _, child := range node.Children {
			if child.Contains3D(&oldBody.Pos.Now) {
				child.insert(oldBody, depth+1)
				found = true
				break
			}
		}
		if !found {
			// This is the case when a body has moved outside of its current
			// node, but hasn't been updated yet. We need to make sure the body
			// remains in a leaf to avoid messing up the references. It doesn't
			// matter which leaf it's in since it's outside of all of them.
			//log.Printf("QuadNode.increaseDepth: %v was not in any children during subdivision. Inserting into first leaf.", oldBody.Entity)
			node.Children[0].addToLeaf(oldBody)
		}
	}
}
func (node *QuadNode) insert(body *Body, depth int) {
	contains := node.Contains3D(&body.Pos.Now)

	if !contains {
		if node.Parent == nil {
			node.expandRoot(&body.Pos.Now)
			node.Tree.Root.insert(body, 0)
		} else {
			log.Printf("QuadNode.insert: body %v is not contained in this node, but also not at root?", body.Entity)
		}
		return
	}

	if node.IsLeaf() {
		if !contains {
			log.Printf("QuadNode.insert: reached leaf, but node doesn't contain body.")
			return
		}
		// Don't recurse too far
		if depth > 8 || len(node.Bodies) < 8 {
			// We can insert here!
			node.addToLeaf(body)
			return
		} else {
			// Full, need to break it up
			node.increaseDepth(depth)
			// Don't return, still need to insert this body into a child node.
		}
	}

	// Not at a leaf, recurse into children
	for _, child := range node.Children {
		if child.Contains3D(&body.Pos.Now) {
			child.insert(body, depth+1)
			return
		}
	}
	log.Printf("QuadNode.insert: tried to insert body %v, but node children don't contain it.", body.Entity)
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

// TODO: this is dumb, should implement https://observablehq.com/@llb4ll/k-nearest-neighbor-search-using-d3-quadtrees
func (node *QuadNode) RangeClosest(center *concepts.Vector3, lightsOnly bool, fn func(b *Body) bool) bool {
	if node.IsLeaf() {
		slice := node.Bodies
		if lightsOnly {
			slice = node.Lights
		}
		// TODO: Should we sort by distance? would require some memory thrashing
		for _, b := range slice {
			if !fn(b) {
				return false
			}
		}
		return true
	}
	cx := (node.Min[0] + node.Max[0]) * 0.5
	cy := (node.Min[1] + node.Max[1]) * 0.5
	first := 0
	if center[0] > cx {
		first = 1
	}
	if center[1] > cy {
		first |= 2
	}
	if !node.Children[first].RangeClosest(center, lightsOnly, fn) {
		return false
	}
	for i := range 4 {
		if i == first {
			continue
		}
		if !node.Children[i].RangeClosest(center, lightsOnly, fn) {
			return false
		}
	}
	return true
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

func (node *QuadNode) aabbOverlaps(min, max *concepts.Vector2) bool {
	if max[0]+node.MaxRadius < node.Min[0] ||
		max[1]+node.MaxRadius < node.Min[1] ||
		min[0]-node.MaxRadius >= node.Max[0] ||
		min[1]-node.MaxRadius >= node.Max[1] {
		return false
	}
	return true
}

func (node *QuadNode) RangeAABB(min, max *concepts.Vector2, fn func(b *Body) bool) {
	if node.IsLeaf() {
		for _, b := range node.Bodies {
			br := b.Size.Now[0] * 0.5
			if b.Pos.Now[0]+br < min[0] ||
				b.Pos.Now[1]+br < min[1] ||
				b.Pos.Now[0]-br >= max[0] ||
				b.Pos.Now[1]-br >= max[1] {
				continue
			}
			if !fn(b) {
				break
			}
		}
		return
	}
	for i := range 4 {
		if !node.Children[i].aabbOverlaps(min, max) {
			continue
		}
		node.Children[i].RangeAABB(min, max, fn)
	}
}
