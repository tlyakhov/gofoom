package logic

func (e *AliveEntity) Hurt(amount float64) {
	e.Health -= amount
}
