package core

import (
	"math"

	"tlyakhov/gofoom/constants"

	"tlyakhov/gofoom/concepts"
)

type Body struct {
	concepts.Attached `editable:"^"`
	Pos               concepts.SimVariable[concepts.Vector3] `editable:"Position"`
	Vel               concepts.SimVariable[concepts.Vector3]
	Size              concepts.SimVariable[concepts.Vector3] `editable:"Size"`
	Force             concepts.Vector3
	Angle             concepts.SimVariable[float64] `editable:"Angle"`
	BoundingRadius    float64                       `editable:"Bounding Radius"`
	Mass              float64                       `editable:"Mass"`
	CollisionResponse CollisionResponse             `editable:"Collision Response"`
	Shadow            BodyShadow                    `editable:"Shadow Type"`
	MountHeight       float64                       `editable:"Mount Height"`
	Active            bool                          `editable:"Active?"`
	OnGround          bool
	SectorEntityRef   *concepts.EntityRef
}

var BodyComponentIndex int

func init() {
	BodyComponentIndex = concepts.DbTypes().Register(Body{}, BodyFromDb)
}

func BodyFromDb(entity *concepts.EntityRef) *Body {
	if asserted, ok := entity.Component(BodyComponentIndex).(*Body); ok {
		return asserted
	}
	return nil
}

func (b *Body) String() string {
	return "Body: " + b.Pos.Now.StringHuman()
}

func (b *Body) SetDB(db *concepts.EntityComponentDB) {
	if b.DB != nil {
		b.Pos.Detach(b.DB.Simulation)
		b.Vel.Detach(b.DB.Simulation)
		b.Size.Detach(b.DB.Simulation)
		b.Angle.Detach(b.DB.Simulation)
	}
	b.Attached.SetDB(db)
	b.Pos.Attach(db.Simulation)
	b.Vel.Attach(db.Simulation)
	b.Size.Attach(db.Simulation)
	b.Angle.Attach(b.DB.Simulation)
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
	b.Pos.Set(concepts.Vector3{0, 0, 0})
	b.Vel.Set(concepts.Vector3{0, 0, 0})
	b.Size.Set(concepts.Vector3{10, 10, 10})
	b.BoundingRadius = 10
	b.CollisionResponse = Slide
	b.MountHeight = constants.PlayerMountHeight
	b.Active = true
	b.Shadow = BodyShadowAABB

	if data == nil {
		return
	}

	if v, ok := data["Active"]; ok {
		b.Active = v.(bool)
	}
	if v, ok := data["Pos"]; ok {
		b.Pos.Deserialize(v)
	}
	if v, ok := data["Vel"]; ok {
		b.Vel.Deserialize(v)
	}
	if v, ok := data["Size"]; ok {
		b.Size.Deserialize(v)
	}
	if v, ok := data["Angle"]; ok {
		b.Angle.Deserialize(v)
	}
	if v, ok := data["BoundingRadius"]; ok {
		b.BoundingRadius = v.(float64)
	}
	if v, ok := data["MountHeight"]; ok {
		b.MountHeight = v.(float64)
	}
	if v, ok := data["Mass"]; ok {
		b.Mass = v.(float64)
	}
	if v, ok := data["CollisionResponse"]; ok {
		c, err := CollisionResponseString(v.(string))
		if err == nil {
			b.CollisionResponse = c
		} else {
			panic(err)
		}
	}
	if v, ok := data["Shadow"]; ok {
		c, err := BodyShadowString(v.(string))
		if err == nil {
			b.Shadow = c
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
	result["Size"] = b.Size.Serialize()
	result["Angle"] = b.Angle.Serialize()
	result["BoundingRadius"] = b.BoundingRadius
	result["Mass"] = b.Mass
	result["MountHeight"] = b.MountHeight
	result["CollisionResponse"] = b.CollisionResponse.String()
	result["Shadow"] = b.Shadow
	return result
}
