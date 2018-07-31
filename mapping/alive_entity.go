package mapping

type AliveEntity struct {
	Entity
	Health   float64
	HurtTime float64
}

func (e *AliveEntity) Initialize() {
	e.Entity.Initialize()
	e.Health = 100
}

func (e *AliveEntity) Deserialize() {

}
