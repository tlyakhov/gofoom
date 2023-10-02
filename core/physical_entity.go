package core

import (
	"math"

	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/registry"

	"tlyakhov/gofoom/concepts"
)

type PhysicalEntity struct {
	*concepts.Base    `editable:"^"`
	Pos               concepts.Vector3 `editable:"Position"`
	Vel               concepts.Vector3
	Angle             float64           `editable:"Angle"`
	BoundingRadius    float64           `editable:"Bounding Radius"`
	Weight            float64           `editable:"Weight"`
	CollisionResponse CollisionResponse `editable:"Collision Response"`
	CRCallback        func() CollisionResponse
	Height            float64 `editable:"Height"`
	MountHeight       float64 `editable:"Mount Height"`
	Active            bool    `editable:"Active?"`
	Sector            AbstractSector
	Map               *Map

	Behaviors map[string]AbstractBehavior `editable:"Behaviors"`
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
	e.Behaviors = make(map[string]AbstractBehavior)
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
		concepts.MapCollection(e, &e.Behaviors, v)
	}
}

func (e *PhysicalEntity) Serialize() map[string]interface{} {
	result := e.Base.Serialize()
	result["Type"] = "core.PhysicalEntity"
	result["Active"] = e.Active
	result["Pos"] = e.Pos.Serialize()
	result["Vel"] = e.Vel.Serialize()
	result["Angle"] = e.Angle
	result["BoundingRadius"] = e.BoundingRadius
	result["Height"] = e.Height
	result["MountHeight"] = e.MountHeight
	result["CollisionResponse"] = e.CollisionResponse.String()

	behaviors := []interface{}{}
	for _, b := range e.Behaviors {
		behaviors = append(behaviors, b.Serialize())
	}
	result["Behaviors"] = behaviors
	return result
}
