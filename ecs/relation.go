package ecs

import (
	"fmt"
	"log"
	"reflect"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/dynamic"
)

type RelationType uint8

const (
	RelationUnknown RelationType = iota
	RelationOne
	RelationSet
	RelationSlice
	RelationTable
)

type Relation struct {
	Owner any
	Name  string
	Type  RelationType
	Value reflect.Value

	One   Entity
	Set   containers.Set[Entity]
	Slice []Entity
	Table EntityTable
}

func (r *Relation) String() string {
	owner := "unknown"
	switch concrete := r.Owner.(type) {
	case fmt.Stringer:
		owner = concrete.String()
	}
	switch r.Type {
	case RelationOne:
		return fmt.Sprintf("Relation [one]: %v.%v = %v", owner, r.Name, r.One.String())
	case RelationSet:
		return fmt.Sprintf("Relation [set]: %v.%v = %v", owner, r.Name, r.Set.String())
	case RelationSlice:
		return fmt.Sprintf("Relation [slice]: %v.%v = %v", owner, r.Name, r.Slice)
	case RelationTable:
		return fmt.Sprintf("Relation [table]: %v.%v = %v", owner, r.Name, r.Table.String())
	}
	return fmt.Sprintf("Relation [unknown]: %v.%v", owner, r.Name)
}

func (r *Relation) Update() {
	if r.Owner == nil || r.Name == "" {
		return
	}

	var updated any
	switch r.Type {
	case RelationOne:
		updated = r.One
	case RelationSet:
		updated = r.Set
	case RelationSlice:
		updated = r.Slice
	case RelationTable:
		updated = r.Table
	}
	r.Value.Set(reflect.ValueOf(updated))
}

func isStructOrPtrToStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct ||
		(t.Kind() == reflect.Pointer &&
			t.Elem().Kind() == reflect.Struct)
}

func relationFromField(field *reflect.StructField, v reflect.Value) Relation {
	r := Relation{
		Value: v,
		Name:  field.Name,
	}

	x := v.Interface()
	switch concreteValue := x.(type) {
	case Entity:
		r.Type = RelationOne
		r.One = concreteValue
	case containers.Set[Entity]:
		r.Type = RelationSet
		r.Set = concreteValue
	case []Entity:
		r.Type = RelationSlice
		r.Slice = concreteValue
	case EntityTable:
		r.Type = RelationTable
		r.Table = concreteValue
	}

	return r
}

func ensurePointerToStruct(v reflect.Value) any {
	if v.Kind() == reflect.Struct {
		return v.Addr().Interface()
	}
	if v.Kind() == reflect.Pointer && v.Elem().Kind() == reflect.Struct {
		return v.Interface()
	}
	return nil
}

const debugRelationWalk = true

// This is a bit ugly and has some fragile aspects:
//  1. The way we walk types is pretty greedy. There's a good chance we're
//     walking into internal caches or unrelated data if they haven't been
//     tagged with ecs:norelation. We should probably decouple the type checking
//     from going through the actual reflected values. A good improvement would
//     be to do the type walk in RegisterComponent, store it in globalTypeMetadata
//     somehow, and then when actually going through an Entity/Component, just
//     walk the cached tree. Then we could debug the relation tree easier, and
//     probably improve performance as well, but there are problems too, for
//     example, how to handle arrays and maps.
//  2. We currently ignore maps, unexported fields, etc...
func rangeComponentRelations(owner any, f func(r *Relation) bool, visited map[any]struct{}, debugPath string) bool {
	if owner == nil {
		return true
	}

	ownerValue := reflect.ValueOf(owner)

	x := ownerValue.Interface()
	switch x.(type) {
	case *Attached:
		// Don't go into the metadata for an attachable component.
		return true
	case dynamic.Spawnable:
		// Dynamics have no relations
		return true
	case Component:
		// If we're at least one level deep, don't go into other components.
		_, isRegistered := Types().IDs[ownerValue.Type().String()]
		if len(visited) > 0 && isRegistered {
			if debugRelationWalk {
				log.Printf("ecs.rangeComponentRelations: %v - skipping component field: %v", debugPath, ownerValue.Type().String())
			}
			return true
		}
	}

	// We've already processed this
	if _, ok := visited[owner]; ok {
		return true
	}
	visited[owner] = struct{}{}

	ownerValue = ownerValue.Elem()
	ownerType := ownerValue.Type()

	if ownerValue.Kind() != reflect.Struct {
		return true
	}

	for i := range ownerType.NumField() {
		field := ownerType.Field(i)

		// Ignore unexported fields or specifically tagged fields.
		if !field.IsExported() || field.Tag.Get("ecs") == "norelation" {
			continue
		}

		if debugRelationWalk {
			log.Printf("ecs.rangeComponentRelations: %v.%v", debugPath, field.Name)
		}

		// Is this field a relation? Run the visitor func.
		r := relationFromField(&field, ownerValue.Field(i))
		if r.Type != RelationUnknown {
			r.Owner = owner
			if debugRelationWalk {
				log.Printf("ecs.rangeComponentRelations: relation found: %v", r.String())
			}
			if !f(&r) {
				return false
			}
			continue
		}

		// Is this field a slice, array, embedded struct we need to recurse into?
		switch r.Value.Kind() {
		case reflect.Slice, reflect.Array:
			if !isStructOrPtrToStruct(field.Type.Elem()) {
				continue
			}
			for i := range r.Value.Len() {
				item := r.Value.Index(i)
				debugPathChild := ""
				if debugRelationWalk {
					debugPathChild = fmt.Sprintf("%v.%v[%v]", debugPath, field.Name, i)
				}
				keepGoing := rangeComponentRelations(ensurePointerToStruct(item), f, visited, debugPathChild)
				if !keepGoing {
					return false
				}
			}
		case reflect.Struct, reflect.Pointer:
			debugPathChild := ""
			if debugRelationWalk {
				debugPathChild = fmt.Sprintf("%v.%v", debugPath, field.Name)
			}
			keepGoing := rangeComponentRelations(ensurePointerToStruct(r.Value), f, visited, debugPathChild)
			if !keepGoing {
				return false
			}
		}
	}
	return true
}
func RangeComponentRelations(owner any, f func(r *Relation) bool) bool {
	debugPath := ""
	if debugRelationWalk {
		debugPath = owner.(Component).Base().Entity.String()
	}
	return rangeComponentRelations(owner, f, make(map[any]struct{}), debugPath)
}

func RangeRelations(e Entity, f func(r *Relation) bool) {
	for _, c := range AllComponents(e) {
		if c == nil {
			continue
		}
		if !RangeComponentRelations(c, f) {
			return
		}
	}
}

func ModifyComponentRelationEntities(owner any, f func(r *Relation, e Entity) Entity) {
	RangeComponentRelations(owner, func(r *Relation) bool {
		switch r.Type {
		case RelationOne:
			updated := f(r, r.One)
			if r.One != updated {
				r.One = updated
				r.Update()
			}
		case RelationSet:
			newSet := make(containers.Set[Entity])
			for e := range r.Set {
				newSet.Add(f(r, e))
			}
			r.Set = newSet
			r.Update()
		case RelationSlice:
			modified := false
			for i, e := range r.Slice {
				updated := f(r, e)
				if updated != e {
					r.Slice[i] = updated
					modified = true
				}
			}
			if modified {
				r.Update()
			}
		case RelationTable:
			newTable := make(EntityTable, len(r.Table))
			for _, e := range r.Table {
				if e == 0 {
					continue
				}
				newTable.Set(f(r, e))
			}
			r.Table = newTable
			r.Update()
		}
		return true
	})
}
