package mapping

import (
	"math"

	"github.com/tlyakhov/gofoom/constants"

	"github.com/tlyakhov/gofoom/concepts"
)

type Entity struct {
	*concepts.Base
	Pos               *concepts.Vector3
	Vel               *concepts.Vector3
	Angle             float64
	BoundingRadius    float64
	CollisionResponse CollisionResponse
	CRCallback        func() CollisionResponse
	Height            float64
	MountHeight       float64
	Active            bool
	Sector            concepts.ISerializable
	Map               *Map
}

func (e *Entity) Initialize() {
	e.Pos = &concepts.Vector3{}
	e.Vel = &concepts.Vector3{}
	e.BoundingRadius = 10
	e.CollisionResponse = Slide
	e.MountHeight = constants.PlayerMountHeight
	e.Active = true
}

func (e *Entity) Angle2DTo(p *concepts.Vector3) float64 {
	dx := e.Pos.X - p.X
	dy := e.Pos.Y - p.Y
	return math.Atan2(dy, dx)*concepts.Rad2deg + 180.0
}

func (e *Entity) SetParent(parent interface{}) {
	if sector, ok := parent.(*Sector); ok {
		e.Sector = sector
		e.Map = sector.Map
	} else {
		panic("Tried mapping.Entity.SetParent with a parameter that wasn't a *mapping.Sector")
	}
}
