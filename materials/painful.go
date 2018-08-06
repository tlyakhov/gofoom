package materials

import "github.com/tlyakhov/gofoom/registry"

type Painful struct {
	Hurt float64 `editable:"Hurt" edit_type:"float"`
}

func init() {
	registry.Instance().Register(Painful{})
}
