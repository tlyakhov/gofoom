package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type EntityEventParams struct {
	Entity ecs.Entity
}

type EntityAxisEventParams struct {
	Entity    ecs.Entity
	AxisValue float64
}

var (
	EventIdForward         = dynamic.RegisterEventClass(&dynamic.EventClass{Name: "Forward"})
	EventIdBack            = dynamic.RegisterEventClass(&dynamic.EventClass{Name: "Back"})
	EventIdLeft            = dynamic.RegisterEventClass(&dynamic.EventClass{Name: "Left"})
	EventIdRight           = dynamic.RegisterEventClass(&dynamic.EventClass{Name: "Right"})
	EventIdPitch           = dynamic.RegisterEventClass(&dynamic.EventClass{Name: "Pitch"})
	EventIdYaw             = dynamic.RegisterEventClass(&dynamic.EventClass{Name: "Yaw"})
	EventIdTurnLeft        = dynamic.RegisterEventClass(&dynamic.EventClass{Name: "TurnLeft"})
	EventIdTurnRight       = dynamic.RegisterEventClass(&dynamic.EventClass{Name: "TurnRight"})
	EventIdUp              = dynamic.RegisterEventClass(&dynamic.EventClass{Name: "Up"})
	EventIdDown            = dynamic.RegisterEventClass(&dynamic.EventClass{Name: "Down"})
	EventIdPrimaryAction   = dynamic.RegisterEventClass(&dynamic.EventClass{Name: "PrimaryAction"})
	EventIdSecondaryAction = dynamic.RegisterEventClass(&dynamic.EventClass{Name: "SecondaryAction"})
)

func init() {
	dynamic.SubscribeToEvent(EventIdForward, eventMove)
	dynamic.SubscribeToEvent(EventIdBack, eventMove)
	dynamic.SubscribeToEvent(EventIdLeft, eventMove)
	dynamic.SubscribeToEvent(EventIdRight, eventMove)
	dynamic.SubscribeToEvent(EventIdTurnLeft, eventTurn)
	dynamic.SubscribeToEvent(EventIdTurnRight, eventTurn)
	dynamic.SubscribeToEvent(EventIdUp, eventUp)
	dynamic.SubscribeToEvent(EventIdDown, eventDown)
	dynamic.SubscribeToEvent(EventIdPrimaryAction, eventPrimaryAction)
	dynamic.SubscribeToEvent(EventIdSecondaryAction, eventSecondaryAction)
	dynamic.SubscribeToEvent(EventIdYaw, eventYaw)
	dynamic.SubscribeToEvent(EventIdPitch, eventPitch)
}

func eventMove(evt *dynamic.Event) bool {
	p := evt.Data.(*EntityEventParams)
	b := core.GetBody(p.Entity)
	if b == nil {
		return false
	}
	switch evt.ID {
	case EventIdForward:
		MovePlayer(p.Entity, b.Angle.Now)
	case EventIdRight:
		MovePlayer(p.Entity, b.Angle.Now+90)
	case EventIdBack:
		MovePlayer(p.Entity, b.Angle.Now+180)
	case EventIdLeft:
		MovePlayer(p.Entity, b.Angle.Now+270)
	}
	return false
}

func eventTurn(evt *dynamic.Event) bool {
	p := evt.Data.(*EntityEventParams)
	b := core.GetBody(p.Entity)
	if b == nil {
		return false
	}
	switch evt.ID {
	case EventIdTurnLeft:
		b.Angle.Now -= constants.PlayerTurnSpeed * constants.TimeStepS
	case EventIdTurnRight:
		b.Angle.Now += constants.PlayerTurnSpeed * constants.TimeStepS
	}
	b.Angle.Now = concepts.NormalizeAngle(b.Angle.Now)
	return false
}

func eventYaw(evt *dynamic.Event) bool {
	p := evt.Data.(*EntityAxisEventParams)
	b := core.GetBody(p.Entity)
	if b == nil {
		return false
	}
	b.Angle.Now = p.AxisValue * 0.25 * constants.PlayerTurnSpeed * constants.TimeStepS
	b.Angle.Now = concepts.NormalizeAngle(b.Angle.Now)
	return false
}

func eventUp(evt *dynamic.Event) bool {
	p := evt.Data.(*EntityEventParams)
	b := core.GetBody(p.Entity)
	m := core.GetMobile(p.Entity)
	if b == nil || m == nil {
		return false
	}
	if behaviors.GetUnderwater(b.SectorEntity) != nil {
		m.Force[2] += constants.PlayerSwimStrength
	} else if b.OnGround {
		m.Force[2] += constants.PlayerJumpForce
		b.OnGround = false
	}
	return false
}

func eventDown(evt *dynamic.Event) bool {
	p := evt.Data.(*EntityEventParams)
	b := core.GetBody(p.Entity)
	m := core.GetMobile(p.Entity)
	player := character.GetPlayer(p.Entity)
	if b == nil || m == nil || player == nil {
		return false
	}
	if behaviors.GetUnderwater(b.SectorEntity) != nil {
		m.Force[2] -= constants.PlayerSwimStrength
	} else {
		player.Crouching = true
	}
	return false
}

func eventPrimaryAction(evt *dynamic.Event) bool {
	p := evt.Data.(*EntityEventParams)
	carrier := inventory.GetCarrier(p.Entity)
	if carrier == nil {
		return false
	}
	if carrier.SelectedWeapon != 0 {
		if w := inventory.GetWeapon(carrier.SelectedWeapon); w != nil {
			w.Intent = inventory.WeaponFire
		}
	}
	return false
}

func eventSecondaryAction(evt *dynamic.Event) bool {
	p := evt.Data.(*EntityEventParams)
	player := character.GetPlayer(p.Entity)
	if player == nil {
		return false
	}
	player.ActionPressed = true
	return false
}

func eventPitch(evt *dynamic.Event) bool {
	p := evt.Data.(*EntityAxisEventParams)
	player := character.GetPlayer(p.Entity)
	if player == nil {
		return false
	}
	player.ShearZ = p.AxisValue * 0.8
	return false
}
