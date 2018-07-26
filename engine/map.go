package engine

import "github.com/tlyakhov/gofoom/util"

type Map struct {
	util.CommonFields

	Sectors []MapSector
	Player  *Entity
}
