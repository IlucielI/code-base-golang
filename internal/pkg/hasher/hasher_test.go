package hasher

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	const pw = "s3cr3t-p@ssw0rd"

	hash, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" || hash == pw {
		t.Fatal("expected a non-empty hash distinct from the plaintext")
	}

	if !VerifyPassword(hash, pw) {
		t.Fatal("expected VerifyPassword to accept the correct password")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Fatal("expected VerifyPassword to reject an incorrect password")
	}
}

func TestHashPassword_ProducesDistinctHashes(t *testing.T) {
	// bcrypt salts each hash, so the same input yields different digests.
	h1, err := HashPassword("same-input")
	if err != nil {
		t.Fatalf("failed to hash first input: %v", err)
	}
	h2, err := HashPassword("same-input")
	if err != nil {
		t.Fatalf("failed to hash second input: %v", err)
	}
	if h1 == h2 {
		t.Fatal("expected bcrypt to produce salted, distinct hashes for equal inputs")
	}
}

func TestHashToken(t *testing.T) {
	got := HashToken("my-token")
	if len(got) != 64 {
		t.Fatalf("expected a 64-char SHA-256 hex digest, got %d chars", len(got))
	}
	if got != HashToken("my-token") {
		t.Fatal("expected HashToken to be deterministic for the same input")
	}
	if got == HashToken("other-token") {
		t.Fatal("expected different tokens to produce different digests")
	}
}
