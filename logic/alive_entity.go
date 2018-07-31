package logic

import (
	"github.com/tlyakhov/gofoom/mapping"
)

type AliveEntity mapping.AliveEntity

func (e *AliveEntity) Hurt(amount float64) {
	e.Health -= amount
}
