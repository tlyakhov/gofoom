package concepts

type Attached struct {
	EntityRef
	Active bool `editable:"Active?"`
}

type Attachable interface {
	Ref() *EntityRef
	SetDB(db *EntityComponentDB)
	Construct(data map[string]any)
	Serialize() map[string]any
}

var AttachedComponentIndex int

func init() {
	AttachedComponentIndex = DbTypes().Register(Attached{})
}

func (a *Attached) Ref() *EntityRef {
	return &a.EntityRef
}

func (a *Attached) SetDB(db *EntityComponentDB) {
	a.EntityRef.DB = db
}

func (a *Attached) Construct(data map[string]any) {
	a.Active = true

	if data == nil {
		return
	}
	if v, ok := data["Entity"]; ok {
		a.Entity = v.(uint64)
	}
	if v, ok := data["Active"]; ok {
		a.Active = v.(bool)
	}
}

func (a *Attached) Serialize() map[string]any {
	return map[string]any{"Entity": a.Entity, "Active": a.Active}
}
