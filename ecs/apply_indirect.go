package ecs

type ApplyIndirect struct {
	ApplyToEntities EntityTable `editable:"Apply To Entities"`
}

func (ai *ApplyIndirect) Construct(data map[string]any) {
	ai.ApplyToEntities = nil

	if data == nil {
		return
	}

	if v, ok := data["ApplyToEntities"]; ok {
		ai.ApplyToEntities = ParseEntityTable(v, true)
	}
}

func (ai *ApplyIndirect) Serialize(result map[string]any) {
	if len(ai.ApplyToEntities) > 0 {
		result["ApplyToEntities"] = ai.ApplyToEntities.Serialize()
	}
}

func (ai *ApplyIndirect) Apply(self Entity, f func(e Entity)) {
	f(self)

	for _, e := range ai.ApplyToEntities {
		if e == 0 {
			continue
		}
		f(e)
	}
}
