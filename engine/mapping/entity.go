package mapping

import "github.com/tlyakhov/gofoom/util"

type Entity struct {
	Pos    util.Vector3
	Angle  float64
	Sector *MapSector
}
