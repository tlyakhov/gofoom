package entity

import (
	"image/color"
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/entities"

	"tlyakhov/gofoom/constants"
)

type PlayerController struct {
	*PhysicalEntityController
	*AliveEntityController
}

func NewPlayerController(p *entities.Player) *PlayerController {
	return &PlayerController{
		PhysicalEntityController: NewPhysicalEntityController(&p.PhysicalEntity),
		AliveEntityController:    NewAliveEntityController(&p.AliveEntity),
	}
}

func (pc *PlayerController) Frame() {
	p := pc.Model.(*entities.Player)
	p.Bob += p.Vel.Now.Length() / 6.0
	for p.Bob > math.Pi*2 {
		p.Bob -= math.Pi * 2
	}
	pc.PhysicalEntityController.Frame()
	if p.Sector == nil {
		return
	}

	if p.Crouching {
		p.Height = constants.PlayerCrouchHeight
	} else {
		p.Height = constants.PlayerHeight
	}

	if p.HurtTime > 0 {
		p.FrameTint = color.NRGBA{0xFF, 0, 0, uint8(p.HurtTime * 200 / constants.PlayerHurtTime)}
		p.HurtTime--
	}
}

func (pc *PlayerController) Hurt(amount float64) {
	p := pc.Model.(*entities.Player)
	pc.AliveEntityController.Hurt(amount)
	p.HurtTime = constants.PlayerHurtTime
}

func (pc *PlayerController) HurtTime() float64 {
	return pc.AliveEntity.HurtTime
}

func (p *PlayerController) Move(angle float64) {
	p.Force[0] += math.Cos(angle*concepts.Deg2rad) * constants.PlayerSpeed * constants.TimeStep
	p.Force[1] += math.Sin(angle*concepts.Deg2rad) * constants.PlayerSpeed * constants.TimeStep
}
