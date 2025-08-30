// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"testing"
)

type mockComponent struct {
	Attached
}

func (*mockComponent) ComponentID() ComponentID {
	return mockCID
}

var mockCID ComponentID

func init() {
	mockCID = RegisterComponent(&Arena[mockComponent, *mockComponent]{})
}

func GetMockComponent(e Entity) *mockComponent {
	if asserted, ok := GetComponent(e, mockCID).(*mockComponent); ok {
		return asserted
	}
	return nil
}

func TestNewEntity(t *testing.T) {
	entity := NewEntity()
	if entity == 0 {
		t.Errorf("NewEntity returned 0")
	}
	if !Entities.Contains(uint32(entity)) {
		t.Errorf("NewEntity did not add entity to Entities")
	}
}

func TestAttach(t *testing.T) {
	entity := NewEntity()
	component := &mockComponent{}
	var a Component = component
	Attach(mockCID, entity, &a)
	if a.Base().Entity != entity {
		t.Errorf("Attach did not set entity")
	}
	if a.Base().Attachments != 1 {
		t.Errorf("Attach did not increase refcount")
	}
}

func TestDetachComponent(t *testing.T) {
	entity := NewEntity()
	component := &mockComponent{}
	var a Component = component
	Attach(mockCID, entity, &a)
	DetachComponent(mockCID, entity)
	if component.Attachments != 0 {
		t.Errorf("DetachComponent did not remove all attachments")
	}
	for _, row := range rows[0] {
		if len(row) > 0 {
			t.Errorf("DetachComponent did not remove the row")
		}
	}
}

func TestDelete(t *testing.T) {
	entity := NewEntity()
	Delete(entity)
	if Entities.Contains(uint32(entity)) {
		t.Errorf("Delete did not remove entity from Entities")
	}
}

func TestComponent(t *testing.T) {
	entity := NewEntity()
	component := NewAttachedComponent(entity, 1)
	if component == nil {
		t.Errorf("Component returned nil")
	}
	if component.ComponentID() != 1 {
		t.Errorf("Component did not have correct ComponentID")
	}
}

func TestSingleton(t *testing.T) {
	component := Singleton(1)
	if component == nil {
		t.Errorf("Singleton returned nil")
	}
	if component.ComponentID() != 1 {
		t.Errorf("Singleton did not have correct ComponentID")
	}
}

func TestAttachTyped(t *testing.T) {
	entity := NewEntity()
	var component *mockComponent
	AttachTyped(entity, &component)
	if component.Entity != entity {
		t.Errorf("AttachTyped did not set entity")
	}
	if !component.IsAttached() {
		t.Errorf("AttachTyped did not set attach")
	}
}

func TestLinked(t *testing.T) {
	entity1 := NewEntity()
	entity2 := NewEntity()
	linked1 := NewAttachedComponent(entity1, LinkedCID).(*Linked)
	linked2 := NewAttachedComponent(entity2, LinkedCID).(*Linked)
	Link(entity1, entity2)
	if !linked1.Entities.Contains(entity2) {
		t.Errorf("Link did not add entity to Entities")
	}
	if !linked2.Entities.Contains(entity1) {
		t.Errorf("Link did not add entity to Entities")
	}
}

func TestNamed(t *testing.T) {
	entity := NewEntity()
	named := NewAttachedComponent(entity, NamedCID).(*Named)
	if named.Name == "" {
		t.Errorf("New name is empty")
	}
	if named.Entity != entity {
		t.Errorf("Named did not set entity")
	}
	if !named.IsAttached() {
		t.Errorf("Named did not set Universe")
	}
}
