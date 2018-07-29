package mapping

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	fmath "github.com/tlyakhov/gofoom/math"
)

type CollisionResponse int

//go:generate stringer -type=CollisionResponse
const (
	Slide CollisionResponse = iota
	Bounce
	Stop
	Remove
	Callback
)

type Entity struct {
	*concepts.Base
	Pos               *fmath.Vector3
	Vel               *fmath.Vector3
	Angle             float64
	BoundingRadius    float64
	CollisionResponse CollisionResponse
	CRCallback        func() CollisionResponse
	Height            float64
	MountHeight       float64
	Active            bool
	Sector            *Sector
	Map               *Map
}

type AliveEntity struct {
	Entity
	Health   float64
	HurtTime float64
}

func (e *Entity) Angle2DTo(p *fmath.Vector3) float64 {
	dx := e.Pos.X - p.X
	dy := e.Pos.Y - p.Y
	return math.Atan2(dy, dx)*fmath.Rad2deg + 180.0
}

func (e *Entity) Remove() {
	if e.Sector != nil {
		delete(e.Sector.Entities, e.ID)
		e.Sector = nil
		return
	}

	for _, item := range e.Map.Sectors {
		if sector, ok := item.(*Sector); ok {
			delete(sector.Entities, e.ID)
		}
	}
}
