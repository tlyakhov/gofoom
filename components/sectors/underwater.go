package sectors

import (
	"tlyakhov/gofoom/concepts"
)

type Underwater struct {
	concepts.Attached `editable:"^"`
}

var UnderwaterComponentIndex int

func init() {
	UnderwaterComponentIndex = concepts.DbTypes().Register(Underwater{})
}

func UnderwaterFromDb(entity *concepts.EntityRef) *Underwater {
	if asserted, ok := entity.Component(UnderwaterComponentIndex).(*Underwater); ok {
		return asserted
	}
	return nil
}
