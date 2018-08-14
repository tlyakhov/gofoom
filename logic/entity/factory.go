package entity

import (
	"sync"

	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/logic/provide"
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
		return NewPhysicalEntityService(target, target)
	case *entities.AliveEntity:
		return NewAliveEntityService(target, target)
	case *entities.Player:
		return NewPlayerService(target)
	case *entities.Light:
		return NewPhysicalEntityService(target.Physical(), target)
	default:
		return nil
		//panic(fmt.Sprintf("Tried to get an entity animator service for %v and didn't find one.", reflect.TypeOf(concrete)))
	}
}

func (f *ColliderFactory) For(concrete interface{}) (provide.Collideable, bool) {
	ea := provide.EntityAnimator.For(concrete)
	if c, ok := ea.(provide.Collideable); ok {
		return c, true
	}

	return nil, false
}

func (f *HurterFactory) For(concrete interface{}) (provide.Hurtable, bool) {
	if concrete == nil {
		return nil, false
	}
	switch target := concrete.(type) {
	case *entities.Player:
		return NewPlayerService(target), true
	case *entities.AliveEntity:
		return NewAliveEntityService(target, target), true
	default:
		//		panic(fmt.Sprintf("Tried to get an entity animator service for %v and didn't find one.", reflect.TypeOf(concrete)))
		return nil, false
	}
}
