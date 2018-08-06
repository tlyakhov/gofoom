package entity

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/tlyakhov/gofoom/logic/provide"
	"github.com/tlyakhov/gofoom/mapping"
)

type AnimatorFactory struct{}

var once sync.Once

func init() {
	once.Do(func() { provide.EntityAnimator = &AnimatorFactory{} })
}

func (f *AnimatorFactory) For(concrete interface{}) provide.Animateable {
	if concrete == nil {
		return nil
	}
	switch target := concrete.(type) {
	case *mapping.PhysicalEntity:
		return NewPhysicalEntityService(target)
	case *mapping.AliveEntity:
		return NewAliveEntityService(target)
	case *mapping.Player:
		return NewPlayerService(target)
	default:
		panic(fmt.Sprintf("Tried to get an entity animator service for %v and didn't find one.", reflect.TypeOf(concrete)))
	}
}
