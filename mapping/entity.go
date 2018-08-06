package mapping

import (
	"math"

	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/registry"

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
	Sector            AbstractSector
	Map               *Map
}

type AbstractEntity interface {
	concepts.ISerializable
	GetEntity() *Entity
	GetSector() AbstractSector
}

func init() {
	registry.Instance().Register(Entity{})
}

func (e *Entity) Initialize() {
	e.Base = &concepts.Base{}
	e.Base.Initialize()
	e.Pos = &concepts.Vector3{}
	e.Vel = &concepts.Vector3{}
	e.BoundingRadius = 10
	e.CollisionResponse = Slide
	e.MountHeight = constants.PlayerMountHeight
	e.Active = true
}

func (e *Entity) GetEntity() *Entity {
	return e
}

func (e *Entity) GetSector() AbstractSector {
	return e.Sector
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
