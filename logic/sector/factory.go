package sector

import (
	"fmt"
	"reflect"
	"sync"

	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/logic/provide"
	"tlyakhov/gofoom/sectors"
)

type AnimatorFactory struct{}
type InteractorFactory struct{}
type PasserFactory struct{}

var once sync.Once

func init() {
	once.Do(func() {
		provide.SectorAnimator = &AnimatorFactory{}
		provide.Interactor = &InteractorFactory{}
		provide.Passer = &PasserFactory{}
	})
}

func (f *InteractorFactory) For(concrete interface{}) provide.Interactable {
	if concrete == nil {
		return nil
	}
	switch target := concrete.(type) {
	case *sectors.ToxicSector:
		return NewToxicSectorService(target)
	case *sectors.VerticalDoor:
		return NewVerticalDoorService(target)
	case *sectors.Underwater:
		return NewUnderwaterService(target)
	case *core.PhysicalSector:
		return NewPhysicalSectorService(target)
	default:
		panic(fmt.Sprintf("Tried to get a sector interactor service for %v and didn't find one.", reflect.TypeOf(concrete)))
	}
}

func (f *AnimatorFactory) For(concrete interface{}) provide.Animateable {
	// For now all sector types are Interactable, Passable, and Animatable, but that may change.
	return provide.Interactor.For(concrete).(provide.Animateable)
}

func (f *PasserFactory) For(concrete interface{}) provide.Passable {
	// For now all sector types are Interactable, Passable, and Animatable, but that may change.
	return provide.Interactor.For(concrete).(provide.Passable)
}
