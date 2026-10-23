package workers

import (
	"context"
	"errors"
	"testing"
)

type samplePayload struct {
	ID string `json:"id"`
}

func TestHandleMessage_Success(t *testing.T) {
	var gotID string
	ack := HandleMessage(context.Background(), "sample", []byte(`{"id":"abc"}`),
		func(_ context.Context, p samplePayload) error {
			gotID = p.ID
			return nil
		})

	if !ack {
		t.Fatal("expected ack=true on successful processing")
	}
	if gotID != "abc" {
		t.Fatalf("expected payload to decode id 'abc', got %q", gotID)
	}
}

func TestHandleMessage_ParseError(t *testing.T) {
	called := false
	ack := HandleMessage(context.Background(), "sample", []byte("not json"),
		func(_ context.Context, _ samplePayload) error {
			called = true
			return nil
		})

	if ack {
		t.Fatal("expected ack=false when payload fails to parse")
	}
	if called {
		t.Fatal("expected process not to run when payload fails to parse")
	}
}

func TestHandleMessage_ProcessError(t *testing.T) {
	ack := HandleMessage(context.Background(), "sample", []byte(`{"id":"abc"}`),
		func(_ context.Context, _ samplePayload) error {
			return errors.New("processing failed")
		})

	if ack {
		t.Fatal("expected ack=false when the process function returns an error")
	}
}

func TestHandleMessage_PropagatesContext(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "value")

	var seen any
	HandleMessage(ctx, "sample", []byte(`{"id":"abc"}`),
		func(c context.Context, _ samplePayload) error {
			seen = c.Value(ctxKey{})
			return nil
		})

	if seen != "value" {
		t.Fatalf("expected broker-provided context to be passed through, got %v", seen)
	}
}
