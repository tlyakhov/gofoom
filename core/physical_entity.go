package core

import (
	"math"

	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/registry"

	"github.com/tlyakhov/gofoom/concepts"
)

type PhysicalEntity struct {
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
	Sector            AbstractSector
	Map               *Map
}

func init() {
	registry.Instance().Register(PhysicalEntity{})
}

func (e *PhysicalEntity) Initialize() {
	e.Base = &concepts.Base{}
	e.Base.Initialize()
	e.Pos = &concepts.Vector3{}
	e.Vel = &concepts.Vector3{}
	e.BoundingRadius = 10
	e.CollisionResponse = Slide
	e.MountHeight = constants.PlayerMountHeight
	e.Active = true
}

func (e *PhysicalEntity) Physical() *PhysicalEntity {
	return e
}

func (e *PhysicalEntity) GetSector() AbstractSector {
	return e.Sector
}

func (e *PhysicalEntity) Angle2DTo(p *concepts.Vector3) float64 {
	dx := e.Pos.X - p.X
	dy := e.Pos.Y - p.Y
	return math.Atan2(dy, dx)*concepts.Rad2deg + 180.0
}

func (e *PhysicalEntity) SetParent(parent interface{}) {
	if sector, ok := parent.(*PhysicalSector); ok {
		e.Sector = sector
		e.Map = sector.Map
	} else {
		panic("Tried mapping.PhysicalEntity.SetParent with a parameter that wasn't a *mapping.PhysicalSector")
	}
}
