package logic

import (
	"reflect"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/mapping"
	"github.com/tlyakhov/gofoom/registry"
)

type Sector mapping.Sector

func init() {
	registry.Instance().RegisterMapped(Sector{}, mapping.Sector{})
}

type IActor interface {
	OnEnter(e *Entity)
	OnExit(e *Entity)
	ActOnEntity(e *Entity)
}

func (s *Sector) OnEnter(e *Entity) {
	eSector := registry.Translate(e.Sector, "logic").(*Sector)
	if s.FloorTarget == nil && e.Pos.Z <= eSector.BottomZ {
		e.Pos.Z = eSector.BottomZ
	}
}

func (s *Sector) OnExit(e *Entity) {
}

func (s *Sector) Collide(e *Entity) {
	entityTop := e.Pos.Z + e.Height

	esector := registry.Translate(e.Sector, "logic").(*Sector)

	if s.FloorTarget != nil && entityTop < s.BottomZ {
		esector.OnExit(e)
		e.Sector = s.FloorTarget
		esector.OnEnter(e)
		e.Pos.Z = esector.TopZ - e.Height - 1.0
	} else if s.FloorTarget == nil && e.Pos.Z <= s.BottomZ {
		e.Vel.Z = 0
		e.Pos.Z = s.BottomZ
	}

	if s.CeilTarget != nil && entityTop > s.TopZ {
		esector.OnExit(e)
		e.Sector = s.CeilTarget
		esector.OnEnter(e)
		e.Pos.Z = esector.BottomZ - e.Height + 1.0
	} else if s.CeilTarget == nil && entityTop > s.TopZ {
		e.Vel.Z = 0
		e.Pos.Z = s.TopZ - e.Height - 1.0
	}
}

func (s *Sector) ActOnEntity(e *Entity) {
	if e.Sector == nil || concepts.ConvertOrCast(e.Sector, reflect.TypeOf(&concepts.Base{})).(*concepts.Base).ID != s.ID {
		return
	}

	if e.ID == s.Map.Player.ID {
		e.Vel.X = 0
		e.Vel.Y = 0
	}

	e.Vel.Z -= constants.Gravity

	s.Collide(e)
}

func (s *Sector) Frame(lastFrameTime float64) {
	for _, item := range s.Entities {
		if e, ok := item.(*Entity); ok {
			if e.ID == s.Map.Player.ID || s.Map.EntitiesPaused {
				continue
			}
			e.Frame(lastFrameTime)
		}
	}
}
