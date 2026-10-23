package queue

import (
	"testing"
)

func TestMustParse(t *testing.T) {
	data := []byte(`{"id":"entity-123"}`)
	var p ID
	MustParse(data, &p)
	if p.ID != "entity-123" {
		t.Errorf("expected entity-123, got %s", p.ID)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected MustParse to panic on invalid json")
		}
	}()
	MustParse([]byte("invalid json"), &p)
}

func TestParse(t *testing.T) {
	data := []byte(`{"id":"order-999"}`)
	var p ID
	if err := Parse(data, &p); err != nil {
		t.Errorf("expected nil error parsing payload, got %v", err)
	}
	if p.ID != "order-999" {
		t.Errorf("expected order-999, got %s", p.ID)
	}

	if err := Parse([]byte("bad json"), &p); err == nil {
		t.Error("expected error parsing invalid json, got nil")
	}
}
