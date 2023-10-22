package mob_controllers

import (
	"image/color"
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/mobs"

	"tlyakhov/gofoom/constants"
)

type PlayerController struct {
	*PhysicalMobController
	*AliveMobController
}

func NewPlayerController(p *mobs.Player) *PlayerController {
	return &PlayerController{
		PhysicalMobController: NewPhysicalMobController(&p.PhysicalMob),
		AliveMobController:    NewAliveMobController(&p.AliveMob),
	}
}

func (pc *PlayerController) Frame() {
	p := pc.Model.(*mobs.Player)
	p.Bob += p.Vel.Now.Length() / 100.0
	for p.Bob > math.Pi*2 {
		p.Bob -= math.Pi * 2
	}
	pc.PhysicalMobController.Frame()
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
	p := pc.Model.(*mobs.Player)
	pc.AliveMobController.Hurt(amount)
	p.HurtTime = constants.PlayerHurtTime
}

func (pc *PlayerController) HurtTime() float64 {
	return pc.AliveMob.HurtTime
}

func (p *PlayerController) Move(angle float64) {
	if p.OnGround {
		p.Force[0] += math.Cos(angle*concepts.Deg2rad) * constants.PlayerWalkForce
		p.Force[1] += math.Sin(angle*concepts.Deg2rad) * constants.PlayerWalkForce
	}
}
