// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"testing"
)

type mockComponent struct {
	Attached
}

var mockCID ComponentID

func init() {
	mockCID = RegisterComponent(&Column[mockComponent, *mockComponent]{Getter: GetMockComponent})
}

func GetMockComponent(db *ECS, e Entity) *mockComponent {
	if asserted, ok := db.Component(e, mockCID).(*mockComponent); ok {
		return asserted
	}
	return nil
}

func TestNewECS(t *testing.T) {
	db := NewECS()
	if db == nil {
		t.Errorf("NewECS returned nil")
	}
}

func TestNewEntity(t *testing.T) {
	db := NewECS()
	entity := db.NewEntity()
	if entity == 0 {
		t.Errorf("NewEntity returned 0")
	}
	if !db.Entities.Contains(uint32(entity)) {
		t.Errorf("NewEntity did not add entity to Entities")
	}
}

func TestAttach(t *testing.T) {
	db := NewECS()
	entity := db.NewEntity()
	component := &mockComponent{}
	var a Attachable = component
	db.Attach(mockCID, entity, &a)
	if a.Base().Entity != entity {
		t.Errorf("Attach did not set entity")
	}
	if a.Base().ECS != db {
		t.Errorf("Attach did not set ECS")
	}
}

func TestDetachComponent(t *testing.T) {
	db := NewECS()
	entity := db.NewEntity()
	component := &mockComponent{}
	var a Attachable = component
	db.Attach(mockCID, entity, &a)
	db.DetachComponent(mockCID, entity)
	if component.Attachments != 0 {
		t.Errorf("DetachComponent did not remove all attachments")
	}
	for _, row := range db.rows {
		if len(row) > 0 {
			t.Errorf("DetachComponent did not remove the row")
		}
	}
}

func TestDelete(t *testing.T) {
	db := NewECS()
	entity := db.NewEntity()
	db.Delete(entity)
	if db.Entities.Contains(uint32(entity)) {
		t.Errorf("Delete did not remove entity from Entities")
	}
}

func TestComponent(t *testing.T) {
	db := NewECS()
	entity := db.NewEntity()
	component := db.NewAttachedComponent(entity, 1)
	if component == nil {
		t.Errorf("Component returned nil")
	}
	if component.Base().ComponentID != 1 {
		t.Errorf("Component did not have correct ComponentID")
	}
}

func TestSingleton(t *testing.T) {
	db := NewECS()
	component := db.Singleton(1)
	if component == nil {
		t.Errorf("Singleton returned nil")
	}
	if component.Base().ComponentID != 1 {
		t.Errorf("Singleton did not have correct ComponentID")
	}
}

func TestAttachTyped(t *testing.T) {
	db := NewECS()
	entity := db.NewEntity()
	var component *mockComponent
	AttachTyped(db, entity, &component)
	if component.Entity != entity {
		t.Errorf("AttachTyped did not set entity")
	}
	if component.ECS != db {
		t.Errorf("AttachTyped did not set ECS")
	}
}

func TestLinked(t *testing.T) {
	db := NewECS()
	entity1 := db.NewEntity()
	entity2 := db.NewEntity()
	linked1 := db.NewAttachedComponent(entity1, LinkedCID).(*Linked)
	linked2 := db.NewAttachedComponent(entity2, LinkedCID).(*Linked)
	db.Link(entity1, entity2)
	if !linked1.Entities.Contains(entity2) {
		t.Errorf("Link did not add entity to Entities")
	}
	if !linked2.Entities.Contains(entity1) {
		t.Errorf("Link did not add entity to Entities")
	}
}

func TestNamed(t *testing.T) {
	db := NewECS()
	entity := db.NewEntity()
	named := db.NewAttachedComponent(entity, NamedCID).(*Named)
	if named.Name == "" {
		t.Errorf("New name is empty")
	}
	if named.Entity != entity {
		t.Errorf("Named did not set entity")
	}
	if named.ECS != db {
		t.Errorf("Named did not set ECS")
	}
}
