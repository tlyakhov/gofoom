package core

import (
	"math"

	"tlyakhov/gofoom/constants"

	"tlyakhov/gofoom/concepts"
)

type Mob struct {
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

var MobComponentIndex int

func init() {
	MobComponentIndex = concepts.DbTypes().Register(Mob{})
}

func MobFromDb(entity *concepts.EntityRef) *Mob {
	return entity.Component(MobComponentIndex).(*Mob)
}

func (e *Mob) SetDB(db *concepts.EntityComponentDB) {
	if e.DB != nil {
		e.Pos.Detach(e.DB.Simulation)
		e.Vel.Detach(e.DB.Simulation)
	}
	e.Attached.SetDB(db)
	e.Pos.Attach(db.Simulation)
	e.Vel.Attach(db.Simulation)
}

func (e *Mob) Sector() *Sector {
	return SectorFromDb(e.SectorEntityRef)
}

func (e *Mob) Angle2DTo(p *concepts.Vector3) float64 {
	dx := e.Pos.Now[0] - p[0]
	dy := e.Pos.Now[1] - p[1]
	return math.Atan2(dy, dx)*concepts.Rad2deg + 180.0
}

func (e *Mob) Construct(data map[string]any) {
	e.Pos.Set(0, 0, 0)
	e.Vel.Set(0, 0, 0)
	e.BoundingRadius = 10
	e.CollisionResponse = Slide
	e.MountHeight = constants.PlayerMountHeight
	e.Active = true

	if data == nil {
		return
	}

	if v, ok := data["Active"]; ok {
		e.Active = v.(bool)
	}
	if v, ok := data["Pos"]; ok {
		e.Pos.Deserialize(v.(map[string]any))
	}
	if v, ok := data["Vel"]; ok {
		e.Vel.Deserialize(v.(map[string]any))
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
}

func (e *Mob) Serialize() map[string]any {
	result := e.Attached.Serialize()
	result["Active"] = e.Active
	result["Pos"] = e.Pos.Serialize()
	result["Vel"] = e.Vel.Serialize()
	result["Angle"] = e.Angle
	result["BoundingRadius"] = e.BoundingRadius
	result["Height"] = e.Height
	result["MountHeight"] = e.MountHeight
	result["CollisionResponse"] = e.CollisionResponse.String()
	return result
}
