package material

import "github.com/tlyakhov/gofoom/mapping"

type IActor interface {
	ActOnEntity(e *mapping.Entity)
}
