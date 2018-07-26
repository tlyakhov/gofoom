package engine

import (
	"math"

	"github.com/tlyakhov/gofoom/util"
)

type Entity struct {
	util.CommonFields

	Pos               *util.Vector3
	Vel               *util.Vector3
	Angle             float64
	HurtTime          float64
	BoundingRadius    float64
	CollisionResponse string // Can be 'slide', 'bounce' 'stop' 'remove' or 'callback'
	Health            float64
	MountHeight       float64
	Active            bool
	Sector            *MapSector
	Map               *Map
}

func (e *Entity) Angle2DTo(p *util.Vector3) float64 {
	dx := e.Pos.X - p.X
	dy := e.Pos.Y - p.Y
	return math.Atan2(dy, dx)*rad2deg + 180.0
}

func (e *Entity) Distance2D(p *util.Vector3) float64 {
	return math.Sqrt((p.X-e.Pos.X)*(p.X-e.Pos.X) + (p.Y-e.Pos.Y)*(p.Y-e.Pos.Y))
}
