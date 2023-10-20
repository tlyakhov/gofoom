package sector

import (
	"fmt"
	"reflect"
	"sync"

	"tlyakhov/gofoom/controllers/provide"
	"tlyakhov/gofoom/core"
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

func (f *InteractorFactory) For(model interface{}) provide.Interactable {
	if model == nil {
		return nil
	}
	switch target := model.(type) {
	case *sectors.ToxicSector:
		return NewToxicSectorController(target)
	case *sectors.VerticalDoor:
		return NewVerticalDoorController(target)
	case *sectors.Underwater:
		return NewUnderwaterController(target)
	case *core.PhysicalSector:
		return NewPhysicalSectorController(target)
	default:
		panic(fmt.Sprintf("Tried to get a sector interactor service for %v and didn't find one.", reflect.TypeOf(model)))
	}
}

func (f *AnimatorFactory) For(model interface{}) provide.Animateable {
	// For now all sector types are Interactable, Passable, and Animatable, but that may change.
	return provide.Interactor.For(model).(provide.Animateable)
}

func (f *PasserFactory) For(model interface{}) provide.Passable {
	// For now all sector types are Interactable, Passable, and Animatable, but that may change.
	return provide.Interactor.For(model).(provide.Passable)
}
