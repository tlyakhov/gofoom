package material

import "github.com/tlyakhov/gofoom/concepts"

type LitSampled struct {
	*concepts.Base
	*Lit
	*Sampled
}

type PainfulLitSampled struct {
	*concepts.Base
	*Lit
	*Sampled
	*Painful
}
