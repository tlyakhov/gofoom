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

func GetMockComponent(u *Universe, e Entity) *mockComponent {
	if asserted, ok := u.Component(e, mockCID).(*mockComponent); ok {
		return asserted
	}
	return nil
}

func TestNewECS(t *testing.T) {
	u := NewUniverse()
	if u == nil {
		t.Errorf("NewECS returned nil")
	}
}

func TestNewEntity(t *testing.T) {
	u := NewUniverse()
	entity := u.NewEntity()
	if entity == 0 {
		t.Errorf("NewEntity returned 0")
	}
	if !u.Entities.Contains(uint32(entity)) {
		t.Errorf("NewEntity did not add entity to Entities")
	}
}

func TestAttach(t *testing.T) {
	u := NewUniverse()
	entity := u.NewEntity()
	component := &mockComponent{}
	var a Attachable = component
	u.Attach(mockCID, entity, &a)
	if a.Base().Entity != entity {
		t.Errorf("Attach did not set entity")
	}
	if a.Base().Universe != u {
		t.Errorf("Attach did not set Universe")
	}
}

func TestDetachComponent(t *testing.T) {
	u := NewUniverse()
	entity := u.NewEntity()
	component := &mockComponent{}
	var a Attachable = component
	u.Attach(mockCID, entity, &a)
	u.DetachComponent(mockCID, entity)
	if component.Attachments != 0 {
		t.Errorf("DetachComponent did not remove all attachments")
	}
	for _, row := range u.rows {
		if len(row) > 0 {
			t.Errorf("DetachComponent did not remove the row")
		}
	}
}

func TestDelete(t *testing.T) {
	u := NewUniverse()
	entity := u.NewEntity()
	u.Delete(entity)
	if u.Entities.Contains(uint32(entity)) {
		t.Errorf("Delete did not remove entity from Entities")
	}
}

func TestComponent(t *testing.T) {
	u := NewUniverse()
	entity := u.NewEntity()
	component := u.NewAttachedComponent(entity, 1)
	if component == nil {
		t.Errorf("Component returned nil")
	}
	if component.Base().ComponentID != 1 {
		t.Errorf("Component did not have correct ComponentID")
	}
}

func TestSingleton(t *testing.T) {
	u := NewUniverse()
	component := u.Singleton(1)
	if component == nil {
		t.Errorf("Singleton returned nil")
	}
	if component.Base().ComponentID != 1 {
		t.Errorf("Singleton did not have correct ComponentID")
	}
}

func TestAttachTyped(t *testing.T) {
	u := NewUniverse()
	entity := u.NewEntity()
	var component *mockComponent
	AttachTyped(u, entity, &component)
	if component.Entity != entity {
		t.Errorf("AttachTyped did not set entity")
	}
	if component.Universe != u {
		t.Errorf("AttachTyped did not set Universe")
	}
}

func TestLinked(t *testing.T) {
	u := NewUniverse()
	entity1 := u.NewEntity()
	entity2 := u.NewEntity()
	linked1 := u.NewAttachedComponent(entity1, LinkedCID).(*Linked)
	linked2 := u.NewAttachedComponent(entity2, LinkedCID).(*Linked)
	u.Link(entity1, entity2)
	if !linked1.Entities.Contains(entity2) {
		t.Errorf("Link did not add entity to Entities")
	}
	if !linked2.Entities.Contains(entity1) {
		t.Errorf("Link did not add entity to Entities")
	}
}

func TestNamed(t *testing.T) {
	u := NewUniverse()
	entity := u.NewEntity()
	named := u.NewAttachedComponent(entity, NamedCID).(*Named)
	if named.Name == "" {
		t.Errorf("New name is empty")
	}
	if named.Entity != entity {
		t.Errorf("Named did not set entity")
	}
	if named.Universe != u {
		t.Errorf("Named did not set Universe")
	}
}
