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

func (f *AnimatorFactory) For(model interface{}) provide.Animateable {
	if model == nil {
		return nil
	}
	switch target := model.(type) {
	case *core.PhysicalEntity:
		return NewPhysicalEntityController(target)
	case *entities.Player:
		return NewPlayerController(target)
	case *entities.Light:
		return NewPhysicalEntityController(target.Physical())
	default:
		return nil
		//panic(fmt.Sprintf("Tried to get an entity animator service for %v and didn't find one.", reflect.TypeOf(model)))
	}
}

func (f *ColliderFactory) For(model interface{}) provide.Collideable {
	if model == nil {
		return nil
	}
	switch target := model.(type) {
	case *core.PhysicalEntity:
		return NewPhysicalEntityController(target)
	case *entities.Player:
		return NewPhysicalEntityController(target.Physical())
	case *entities.Light:
		return NewPhysicalEntityController(target.Physical())
	default:
		return nil
		//panic(fmt.Sprintf("Tried to get an collider service for %v and didn't find one.", reflect.TypeOf(model)))
	}
}

func (f *HurterFactory) For(model interface{}) provide.Hurtable {
	if model == nil {
		return nil
	}
	switch target := model.(type) {
	case *entities.Player:
		return NewPlayerController(target)
	case *entities.AliveEntity:
		return NewAliveEntityController(target)
	default:
		//		panic(fmt.Sprintf("Tried to get an entity animator service for %v and didn't find one.", reflect.TypeOf(model)))
		return nil
	}
}
