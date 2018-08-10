package core

import (
	"math"

	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/registry"

	"github.com/tlyakhov/gofoom/concepts"
)

type PhysicalEntity struct {
	*concepts.Base
	Pos               concepts.Vector3
	Vel               concepts.Vector3
	Angle             float64
	BoundingRadius    float64
	CollisionResponse CollisionResponse
	CRCallback        func() CollisionResponse
	Height            float64
	MountHeight       float64
	Active            bool
	Sector            AbstractSector
	Map               *Map

	behaviors map[string]AbstractBehavior
}

func init() {
	registry.Instance().Register(PhysicalEntity{})
}

func (e *PhysicalEntity) Initialize() {
	e.Base = &concepts.Base{}
	e.Base.Initialize()
	e.Pos = concepts.Vector3{}
	e.Vel = concepts.Vector3{}
	e.BoundingRadius = 10
	e.CollisionResponse = Slide
	e.MountHeight = constants.PlayerMountHeight
	e.Active = true
	e.behaviors = make(map[string]AbstractBehavior)
}

func (e *PhysicalEntity) Physical() *PhysicalEntity {
	return e
}

func (e *PhysicalEntity) Behaviors() map[string]AbstractBehavior {
	return e.behaviors
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
	if sector, ok := parent.(AbstractSector); ok {
		e.Sector = sector
		e.Map = sector.Physical().Map
	} else {
		panic("Tried mapping.PhysicalEntity.SetParent with a parameter that wasn't a *mapping.PhysicalSector")
	}
}

func (e *PhysicalEntity) Deserialize(data map[string]interface{}) {
	e.Initialize()
	e.Base.Deserialize(data)

	if v, ok := data["Active"]; ok {
		e.Active = v.(bool)
	}
	if v, ok := data["Pos"]; ok {
		e.Pos.Deserialize(v.(map[string]interface{}))
	}
	if v, ok := data["Vel"]; ok {
		e.Vel.Deserialize(v.(map[string]interface{}))
	}
	if v, ok := data["Angle"]; ok {
		e.Angle = v.(float64)
	}
	if v, ok := data["BoundingRadius"]; ok {
		e.BoundingRadius = v.(float64)
	}
	if v, ok := data["Height"]; ok {
		e.Height = v.(float64)
	}
	if v, ok := data["MountHeight"]; ok {
		e.MountHeight = v.(float64)
	}
	if v, ok := data["CollisionResponse"]; ok {
		c, err := CollisionResponseString(v.(string))
		if err == nil {
			e.CollisionResponse = c
		} else {
			panic(err)
		}
	}
	if v, ok := data["Behaviors"]; ok {
		concepts.MapCollection(e, &e.behaviors, v)
	}
}
