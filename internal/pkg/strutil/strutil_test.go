package strutil

import "testing"

func TestConstantTimeEqual(t *testing.T) {
	if !ConstantTimeEqual("abc123", "abc123") {
		t.Fatal("expected equal strings to compare true")
	}
	if ConstantTimeEqual("abc123", "abc124") {
		t.Fatal("expected different strings to compare false")
	}
	if ConstantTimeEqual("abc", "abcd") {
		t.Fatal("expected different-length strings to compare false")
	}
}

func TestLikePattern(t *testing.T) {
	cases := map[string]string{
		"hello":    "%hello%",
		"  spaced ": "%spaced%",
		"":          "",
		"   ":       "",
	}
	for in, want := range cases {
		if got := LikePattern(in); got != want {
			t.Errorf("LikePattern(%q) = %q, want %q", in, got, want)
		}
	}
}
