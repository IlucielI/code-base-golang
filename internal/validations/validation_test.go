package validations

import (
	"errors"
	"testing"
)

type dummyValidatable struct {
	valid bool
}

func (d dummyValidatable) Validate() error {
	if !d.valid {
		return errors.New("invalid")
	}
	return nil
}

func TestValidate(t *testing.T) {
	if err := Validate(nil); err != nil {
		t.Fatalf("expected nil error for nil validatable, got %v", err)
	}

	if err := Validate(dummyValidatable{valid: true}); err != nil {
		t.Fatalf("expected nil error for valid struct, got %v", err)
	}

	if err := Validate(dummyValidatable{valid: false}); err == nil {
		t.Fatal("expected error for invalid struct, got nil")
	}
}
