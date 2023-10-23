package concepts

type Attached struct {
	Entity uint64
	DB     *EntityComponentDB
	Active bool `editable:"Active?"`
}

type Attachable interface {
	GetEntity() uint64
	SetEntity(uint64)
	GetDB() *EntityComponentDB
	SetDB(db *EntityComponentDB)
	EntityRef() *EntityRef
	Construct(data map[string]any)
	Serialize() map[string]any
}

var AttachedComponentIndex int

func init() {
	AttachedComponentIndex = DbTypes().Register(Attached{})
}

func (a *Attached) GetEntity() uint64 {
	return a.Entity
}

func (a *Attached) SetEntity(e uint64) {
	a.Entity = e
}

func (a *Attached) GetDB() *EntityComponentDB {
	return a.DB
}

func (a *Attached) SetDB(db *EntityComponentDB) {
	a.DB = db
}

func (a *Attached) EntityRef() *EntityRef {
	return &EntityRef{DB: a.DB, Entity: a.Entity}
}

func (a *Attached) Construct(data map[string]any) {
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
