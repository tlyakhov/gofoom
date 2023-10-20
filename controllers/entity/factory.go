package entity

import (
	"sync"

	"tlyakhov/gofoom/controllers/provide"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/entities"
)

type AnimatorFactory struct{}
type HurterFactory struct{}
type ColliderFactory struct{}

var once sync.Once

func init() {
	once.Do(func() {
		provide.EntityAnimator = &AnimatorFactory{}
		provide.Hurter = &HurterFactory{}
		provide.Collider = &ColliderFactory{}
	})
}

func (f *AnimatorFactory) For(concrete interface{}) provide.Animateable {
	if concrete == nil {
		return nil
	}
	switch target := concrete.(type) {
	case *core.PhysicalEntity:
		return NewPhysicalEntityController(target, target)
	case *entities.AliveEntity:
		return NewAliveEntityController(target, target)
	case *entities.Player:
		return NewPlayerController(target)
	case *entities.Light:
		return NewPhysicalEntityController(target.Physical(), target)
	default:
		return nil
		//panic(fmt.Sprintf("Tried to get an entity animator service for %v and didn't find one.", reflect.TypeOf(concrete)))
	}
}

func (f *ColliderFactory) For(concrete interface{}) (provide.Collideable, bool) {
	if concrete == nil {
		return nil, false
	}
	switch target := concrete.(type) {
	case *core.PhysicalEntity:
		return NewPhysicalEntityController(target, target), true
	case *entities.AliveEntity:
		return NewAliveEntityController(target, target), true
	case *entities.Player:
		return NewPlayerController(target), true
	case *entities.Light:
		return NewPhysicalEntityController(target.Physical(), target), true
	default:
		return nil, false
		//panic(fmt.Sprintf("Tried to get an collider service for %v and didn't find one.", reflect.TypeOf(concrete)))
	}
}

func (f *HurterFactory) For(concrete interface{}) (provide.Hurtable, bool) {
	if concrete == nil {
		return nil, false
	}
	switch target := concrete.(type) {
	case *entities.Player:
		return NewPlayerController(target), true
	case *entities.AliveEntity:
		return NewAliveEntityController(target, target), true
	default:
		//		panic(fmt.Sprintf("Tried to get an entity animator service for %v and didn't find one.", reflect.TypeOf(concrete)))
		return nil, false
	}
}
