package concepts

import (
	"strconv"
	"strings"
)

type Attached struct {
	*EntityRef `editable:"Component" edit_type:"Component"`
	Active     bool `editable:"Active?"`
	indexInDB  int
}

type Attachable interface {
	Serializable
	Ref() *EntityRef
	ResetRef()
	SetDB(db *EntityComponentDB)
	String() string
	IndexInDB() int
	SetIndexInDB(int)
}

type Serializable interface {
	Construct(data map[string]any)
	Serialize() map[string]any
}

var AttachedComponentIndex int

func init() {
	AttachedComponentIndex = DbTypes().Register(Attached{}, nil)
}

func (a *Attached) String() string {
	return a.EntityRef.String()
}

func (a *Attached) Ref() *EntityRef {
	return a.EntityRef
}

func (a *Attached) IndexInDB() int {
	return a.indexInDB
}

func (a *Attached) SetIndexInDB(i int) {
	a.indexInDB = i
}

func (a *Attached) ResetRef() {
	var db *EntityComponentDB
	if a.EntityRef != nil {
		db = a.EntityRef.DB
	}
	a.EntityRef = &EntityRef{DB: db}
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
		a.Entity, _ = strconv.ParseUint(v.(string), 10, 64)
	}
	if v, ok := data["Active"]; ok {
		a.Active = v.(bool)
	}
}

func (a *Attached) Serialize() map[string]any {
	return map[string]any{"Entity": strconv.FormatUint(a.Entity, 10), "Active": a.Active}
}

func (a *Attached) DeserializeComponentList(list *map[int]bool, name string, data map[string]any) {
	v, ok := data[name]
	if !ok {
		return
	}
	listString, ok := v.(string)
	if !ok {
		return
	}
	split := strings.Split(listString, ",")
	*list = make(map[int]bool)
	for _, typeName := range split {
		componentIndex := DbTypes().Indexes[typeName]
		if componentIndex != 0 {
			(*list)[componentIndex] = true
		}
	}
}

func (a *Attached) SerializeComponentList(list map[int]bool, name string, result map[string]any) {
	if len(list) == 0 {
		return
	}

	types := make([]string, 0)
	for index := range list {
		types = append(types, DbTypes().Types[index].String())
	}
	result[name] = strings.Join(types, ",")
}
