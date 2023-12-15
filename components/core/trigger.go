package core

import "tlyakhov/gofoom/concepts"

type Trigger struct {
	Condition Script `editable:"Condition"`
	Action    Script `editable:"Action"`
}

func (t *Trigger) Construct(db *concepts.EntityComponentDB, data map[string]any) {
	t.Condition.Construct(db, nil)
	t.Action.Construct(db, nil)

	if data == nil {
		return
	}

	if v, ok := data["Condition"]; ok {
		t.Condition.Construct(db, v.(map[string]any))
	}
	if v, ok := data["Action"]; ok {
		t.Action.Construct(db, v.(map[string]any))
	}
}

func (t *Trigger) Serialize() map[string]any {
	result := make(map[string]any)
	result["Condition"] = t.Condition.Serialize()
	result["Action"] = t.Action.Serialize()

	return result
}

func ConstructTriggers(db *concepts.EntityComponentDB, data any) []Trigger {
	var result []Trigger

	if triggers, ok := data.([]any); ok {
		result = make([]Trigger, len(triggers))
		for i, tdata := range triggers {
			result[i].Construct(db, tdata.(map[string]any))
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
