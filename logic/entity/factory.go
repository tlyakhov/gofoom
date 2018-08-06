package entity

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/logic/provide"
)

type AnimatorFactory struct{}
type HurterFactory struct{}

var once sync.Once

func init() {
	once.Do(func() {
		provide.EntityAnimator = &AnimatorFactory{}
		provide.Hurter = &HurterFactory{}
	})
}

func (f *AnimatorFactory) For(concrete interface{}) provide.Animateable {
	if concrete == nil {
		return nil
	}
	switch target := concrete.(type) {
	case *core.PhysicalEntity:
		return NewPhysicalEntityService(target)
	case *entities.AliveEntity:
		return NewAliveEntityService(target)
	case *entities.Player:
		return NewPlayerService(target)
	default:
		panic(fmt.Sprintf("Tried to get an entity animator service for %v and didn't find one.", reflect.TypeOf(concrete)))
	}
}

func (f *HurterFactory) For(concrete interface{}) provide.Hurtable {
	if concrete == nil {
		return nil
	}
	switch target := concrete.(type) {
	case *entities.AliveEntity:
		return NewAliveEntityService(target)
	case *entities.Player:
		return NewPlayerService(target)
	default:
		panic(fmt.Sprintf("Tried to get an entity animator service for %v and didn't find one.", reflect.TypeOf(concrete)))
	}
}
