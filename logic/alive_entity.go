package logic

import (
	"github.com/tlyakhov/gofoom/mapping"
	"github.com/tlyakhov/gofoom/registry"
)

type AliveEntity mapping.AliveEntity

func init() {
	registry.Instance().RegisterMapped(AliveEntity{}, mapping.AliveEntity{})
}

func (e *AliveEntity) Hurt(amount float64) {
	e.Health -= amount
}
