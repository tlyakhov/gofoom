package mob_controllers

import (
	"sync"

	"tlyakhov/gofoom/controllers/provide"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/mobs"
)

type AnimatorFactory struct{}
type HurterFactory struct{}
type ColliderFactory struct{}

var once sync.Once

func init() {
	once.Do(func() {
		provide.MobAnimator = &AnimatorFactory{}
		provide.Hurter = &HurterFactory{}
		provide.Collider = &ColliderFactory{}
	})
}

func (f *AnimatorFactory) For(model interface{}) provide.Animateable {
	if model == nil {
		return nil
	}
	switch target := model.(type) {
	case *core.PhysicalMob:
		return NewPhysicalMobController(target)
	case *mobs.Player:
		return NewPlayerController(target)
	case *mobs.Light:
		return NewPhysicalMobController(target.Physical())
	default:
		return nil
		//panic(fmt.Sprintf("Tried to get an mob animator service for %v and didn't find one.", reflect.TypeOf(model)))
	}
}

func (f *ColliderFactory) For(model interface{}) provide.Collideable {
	if model == nil {
		return nil
	}
	switch target := model.(type) {
	case *core.PhysicalMob:
		return NewPhysicalMobController(target)
	case *mobs.Player:
		return NewPhysicalMobController(target.Physical())
	case *mobs.Light:
		return NewPhysicalMobController(target.Physical())
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
	case *mobs.Player:
		return NewPlayerController(target)
	case *mobs.AliveMob:
		return NewAliveMobController(target)
	default:
		//		panic(fmt.Sprintf("Tried to get an mob animator service for %v and didn't find one.", reflect.TypeOf(model)))
		return nil
	}
}
