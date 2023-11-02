package core

import (
	"math"

	"tlyakhov/gofoom/constants"

	"tlyakhov/gofoom/concepts"
)

type Body struct {
	concepts.Attached `editable:"^"`
	Pos               concepts.SimVector3 `editable:"Position"`
	Vel               concepts.SimVector3
	Force             concepts.Vector3
	Angle             float64           `editable:"Angle"`
	BoundingRadius    float64           `editable:"Bounding Radius"`
	Mass              float64           `editable:"Mass"`
	CollisionResponse CollisionResponse `editable:"Collision Response"`
	Height            float64           `editable:"Height"`
	MountHeight       float64           `editable:"Mount Height"`
	Active            bool              `editable:"Active?"`
	OnGround          bool
	SectorEntityRef   *concepts.EntityRef
}

var BodyComponentIndex int

func init() {
	BodyComponentIndex = concepts.DbTypes().Register(Body{})
}

func BodyFromDb(entity *concepts.EntityRef) *Body {
	if asserted, ok := entity.Component(BodyComponentIndex).(*Body); ok {
		return asserted
	}
	return nil
}

func (b *Body) SetDB(db *concepts.EntityComponentDB) {
	if b.DB != nil {
		b.Pos.Detach(b.DB.Simulation)
		b.Vel.Detach(b.DB.Simulation)
	}
	b.Attached.SetDB(db)
	b.Pos.Attach(db.Simulation)
	b.Vel.Attach(db.Simulation)
}

func (b *Body) Sector() *Sector {
	return SectorFromDb(b.SectorEntityRef)
}

func (b *Body) Angle2DTo(p *concepts.Vector3) float64 {
	dx := b.Pos.Now[0] - p[0]
	dy := b.Pos.Now[1] - p[1]
	return math.Atan2(dy, dx)*concepts.Rad2deg + 180.0
}

func (b *Body) Construct(data map[string]any) {
	b.Pos.Set(0, 0, 0)
	b.Vel.Set(0, 0, 0)
	b.BoundingRadius = 10
	b.CollisionResponse = Slide
	b.MountHeight = constants.PlayerMountHeight
	b.Active = true

	if data == nil {
		return
	}

	if v, ok := data["Active"]; ok {
		b.Active = v.(bool)
	}
	if v, ok := data["Pos"]; ok {
		b.Pos.Deserialize(v.(map[string]any))
	}
	if v, ok := data["Vel"]; ok {
		b.Vel.Deserialize(v.(map[string]any))
	}
	if v, ok := data["Angle"]; ok {
		b.Angle = v.(float64)
	}
	if v, ok := data["BoundingRadius"]; ok {
		b.BoundingRadius = v.(float64)
	}
	if v, ok := data["Height"]; ok {
		b.Height = v.(float64)
	}
	if v, ok := data["MountHeight"]; ok {
		b.MountHeight = v.(float64)
	}
	if v, ok := data["CollisionResponse"]; ok {
		c, err := CollisionResponseString(v.(string))
		if err == nil {
			b.CollisionResponse = c
		} else {
			panic(err)
		}
	}
}

func (b *Body) Serialize() map[string]any {
	result := b.Attached.Serialize()
	result["Active"] = b.Active
	result["Pos"] = b.Pos.Serialize()
	result["Vel"] = b.Vel.Serialize()
	result["Angle"] = b.Angle
	result["BoundingRadius"] = b.BoundingRadius
	result["Height"] = b.Height
	result["MountHeight"] = b.MountHeight
	result["CollisionResponse"] = b.CollisionResponse.String()
	return result
}
