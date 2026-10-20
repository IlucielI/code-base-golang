package jsonutil

import "testing"

func TestRawStringSlice(t *testing.T) {
	if got := string(RawStringSlice(nil)); got != "[]" {
		t.Errorf("nil slice: got %s, want []", got)
	}
	if got := string(RawStringSlice([]string{})); got != "[]" {
		t.Errorf("empty slice: got %s, want []", got)
	}
	if got := string(RawStringSlice([]string{"a", "b"})); got != `["a","b"]` {
		t.Errorf(`["a","b"]: got %s`, got)
	}
}
