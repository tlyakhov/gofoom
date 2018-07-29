package logic

import "github.com/tlyakhov/gofoom/constants"

func (s *Sector) OnEnter(e *Entity) {
	if s.FloorTarget == nil && e.Pos.Z <= e.Sector.BottomZ {
		e.Pos.Z = e.Sector.BottomZ
	}
}

func (s *Sector) OnExit(e *Entity) {
}

func (s *Sector) Collide(e *Entity) {
	entityTop := e.Pos.Z + e.Height

	if ms.FloorTarget != nil && entityTop < ms.BottomZ {
		e.Sector.OnExit(e)
		e.Sector = ms.FloorTarget
		e.Sector.OnEnter(e)
		e.Pos.Z = e.Sector.TopZ - e.Height - 1.0
	} else if ms.FloorTarget == nil && e.Pos.Z <= ms.BottomZ {
		e.Vel.Z = 0
		e.Pos.Z = ms.BottomZ
	}

	if ms.CeilTarget != nil && entityTop > ms.TopZ {
		e.Sector.OnExit(e)
		e.Sector = ms.CeilTarget
		e.Sector.OnEnter(e)
		e.Pos.Z = e.Sector.BottomZ - e.Height + 1.0
	} else if ms.CeilTarget == nil && entityTop > ms.TopZ {
		e.Vel.Z = 0
		e.Pos.Z = ms.TopZ - e.Height - 1.0
	}
}

func (s *BasicSector) Collide(e Entity) {
	ms := s.Sector()
	me := e.MapEntity()
	ms.Collide(me)

	if ae, ok := e.(*AliveEntity); ok && s.Hurt != 0 && ae.HurtTime == 0 {
		ae.Hurt(s.Hurt)
	}

	if ms.FloorMaterial != nil && me.Pos.Z <= ms.BottomZ {
		ms.FloorMaterial.ActOnEntity(e)
	}
	if ms.CeilMaterial != nil && me.Pos.Z >= ms.TopZ {
		ms.CeilMaterial.ActOnEntity(e)
	}
}

func (ms *Sector) ActOnEntity(e *Entity) {
	if e.Sector == nil || e.Sector.ID != ms.ID {
		return
	}

	if e.ID == ms.Map.Player.ID {
		e.Vel.X = 0
		e.Vel.Y = 0
	}

	e.Vel.Z -= constants.Gravity

	ms.Collide(e)

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
