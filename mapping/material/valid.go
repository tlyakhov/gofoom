package material

var (
	ValidMaterialTypes = map[string]interface{}{
		"material.Lit":               Lit{},
		"material.Sampled":           Sampled{},
		"material.Sky":               Sky{},
		"material.LitSampled":        LitSampled{},
		"material.PainfulLitSampled": PainfulLitSampled{},
	}
)
