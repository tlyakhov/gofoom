package core

type Trigger struct {
	Condition Expression `editable:"Condition"`
	Action    Expression `editable:"Action"`
}

func (t *Trigger) Construct(data map[string]any) {
	if data == nil {
		return
	}

	if v, ok := data["Condition"]; ok {
		t.Condition.Construct(v.(string))
	}
	if v, ok := data["Action"]; ok {
		t.Action.Construct(v.(string))
	}
}

func (t *Trigger) Serialize() map[string]any {
	result := make(map[string]any)
	result["Condition"] = t.Condition.Serialize()
	result["Action"] = t.Action.Serialize()

	return result
}

func ConstructTriggers(data any) []Trigger {
	var result []Trigger

	if triggers, ok := data.([]any); ok {
		result = make([]Trigger, len(triggers))
		for i, tdata := range triggers {
			result[i].Construct(tdata.(map[string]any))
		}
	}
	return result
}

func SerializeTriggers(triggers []Trigger) []map[string]any {
	result := make([]map[string]any, len(triggers))
	for i, trigger := range triggers {
		result[i] = trigger.Serialize()
	}
	return result
}
