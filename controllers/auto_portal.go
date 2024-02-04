package controllers

import (
	"log"
	"math"
	"strconv"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

// TODO: Fix situations where sectors are on top of each other and shouldn't be
// connected up
func autoCheckSegment(a, b *core.Segment) bool {
	/* We have multiple cases:
	1. The segments match, in which case their adjacencies should be wired up.
	2. Segments a & b are not collinear
	3. Segment a and b don't overlap
	4. Segment b encloses segment a
	5. Segment a encloses segment b */

	// Case 1:
	if a.Matches(b) {
		b.AdjacentSector = a.Sector.Ref()
		b.AdjacentSegment = a
		a.AdjacentSector = b.Sector.Ref()
		a.AdjacentSegment = b
		return false
	}

	// Check for co-linearity
	aDelta := &concepts.Vector2{a.P[0] - a.Next.P[0], a.P[1] - a.Next.P[1]}
	bDelta := &concepts.Vector2{a.P[0] - b.Next.P[0], a.P[1] - b.Next.P[1]}
	abDelta := &concepts.Vector2{a.P[0] - b.P[0], a.P[1] - b.P[1]}
	c1 := aDelta.Cross(bDelta)
	c2 := aDelta.Cross(abDelta)

	if c1 < -constants.IntersectEpsilon || c1 > constants.IntersectEpsilon ||
		c2 < -constants.IntersectEpsilon || c2 > constants.IntersectEpsilon {
		// Case 2
		return false
	}

	split := false
	/*log.Printf("[%v] (%v)<->(%v) is co-linear with [%v] (%v)<->(%v)\n",
	a.Sector.Entity, a.P.StringHuman(), a.Next.P.StringHuman(),
	b.Sector.Entity, b.P.StringHuman(), b.Next.P.StringHuman())*/

	// which axis should we use for comparisons?
	xRange := math.Max(math.Abs(a.P[0]-a.Next.P[0]), math.Abs(b.P[0]-b.Next.P[0]))
	yRange := math.Max(math.Abs(a.P[1]-a.Next.P[1]), math.Abs(b.P[1]-b.Next.P[1]))
	axis := 0
	if yRange > xRange {
		axis = 1
	}
	a1 := a.P
	a2 := a.Next.P
	b1 := b.P
	b2 := b.Next.P

	// Ensure the comparison order is the same for both segments
	aSwap := false
	if a2[axis] < a1[axis] {
		a1, a2 = a2, a1
		aSwap = true
	}

	bSwap := false
	if b2[axis] < b1[axis] {
		b1, b2 = b2, b1
		bSwap = true
	}

	// Do we need to do anything to segment A?
	//    a1-----b1------------a2
	if b1[axis] > a1[axis] && b1[axis] < a2[axis] {
		split = true
		log.Printf("Splitting segment A by B")

		aSplit := a.Split(b1)

		// Now we have a couple of cases:
		// 1. b2 could be between b1 and a2
		//    a1-----b1-----b2-----a2
		// 2. b2 could be > a2 (so we do nothing)
		//    a1-----b1------------a2-----b2
		if b2[axis] < a2[axis] {
			log.Printf("Splitting the second half of split A by B.Next")
			if aSwap {
				a.Split(b2)
			} else {
				aSplit.Split(b2)
			}
		}
	} else if b2[axis] > a1[axis] && b2[axis] < a2[axis] {
		//    b1-----a1-----b2-----a2
		log.Printf("Splitting segment A by B.Next")
		split = true
		a.Split(b2)
	}

	// Do we need to do anything to segment B?
	//    b1-----a1------------b2
	if a1[axis] > b1[axis] && a1[axis] < b2[axis] {
		log.Printf("Splitting segment B by A")
		split = true

		bSplit := b.Split(a1)
		// Now we have a couple of cases:
		// 1. a2 could be between a1 and b2
		//    b1-----a1-----a2-----b2
		// 2. a2 could be > b2 (so we do nothing)
		//    b1-----a1------------b2-----a2
		if a2[axis] < b2[axis] {
			log.Printf("Splitting the second half of split B by A.Next")
			if bSwap {
				b.Split(a2)
			} else {
				bSplit.Split(a2)
			}
		}
	} else if a2[axis] > b1[axis] && a2[axis] < b2[axis] {
		//    a1-----b1-----a2-----b2
		log.Printf("Splitting segment B by A.Next")
		split = true
		b.Split(a2)
	}
	return split
}

// Automatically connects adjacent sectors, potentially splitting segments
// as necessary. This is a really expensive function O(n^2 + n), although for
// most cases each comparison is pretty cheap. For really large worlds, we could
// give the user the option of disabling auto-portalling in the editor and only
// doing it manually every once in a while.
func AutoPortal(db *concepts.EntityComponentDB) {
	seen := map[string]bool{}
	for _, c := range db.All(core.SectorComponentIndex) {
		sector := c.(*core.Sector)
		for _, segment := range sector.Segments {
			segment.AdjacentSector = nil
			segment.AdjacentSegment = nil
		}
	}
	for _, c := range db.All(core.SectorComponentIndex) {
		for _, c2 := range db.All(core.SectorComponentIndex) {
			if c == c2 {
				continue
			}
			name := strconv.FormatUint(c.Ref().Entity, 10) + "|" + strconv.FormatUint(c2.Ref().Entity, 10)
			id2 := strconv.FormatUint(c2.Ref().Entity, 10) + "|" + strconv.FormatUint(c.Ref().Entity, 10)
			if seen[id2] || seen[name] {
				continue
			}
			seen[name] = true

			sector := c.(*core.Sector)
			sector2 := c2.(*core.Sector)
			if !sector.AABBIntersect(&sector2.Min, &sector2.Max, true) {
				continue
			}

			split := true
			for split {
				for _, segment := range sector.Segments {
					for _, segment2 := range sector2.Segments {
						split = autoCheckSegment(segment, segment2)
						if split {
							break
						}
					}
					if split {
						break
					}
				}
			}
		}
	}
}
