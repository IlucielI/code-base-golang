package apperror

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	// Without raw error
	err1 := New(http.StatusBadRequest, "BAD_REQUEST", "invalid title")
	if err1.Error() != "invalid title" {
		t.Fatalf("expected 'invalid title', got '%s'", err1.Error())
	}

	// With raw error
	raw := errors.New("underlying db crash")
	err2 := New(http.StatusInternalServerError, "INTERNAL_ERROR", "database error", raw)
	if err2.Error() != "database error" {
		t.Fatalf("unexpected error string: %s", err2.Error())
	}
	if !errors.Is(err2, raw) {
		t.Fatal("expected errors.Is to match raw error through Unwrap")
	}

	// With empty message but raw error
	err3 := New(http.StatusInternalServerError, "INTERNAL_ERROR", "", raw)
	if err3.Error() != "underlying db crash" {
		t.Fatalf("expected raw error message, got: %s", err3.Error())
	}

	// Nil error receiver
	var nilErr *AppError
	if nilErr.Error() != "" {
		t.Fatalf("expected empty string for nil error, got '%s'", nilErr.Error())
	}
	if nilErr.Unwrap() != nil {
		t.Fatal("expected nil unwrap for nil error")
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := New(http.StatusNotFound, "NOT_FOUND", "task not found")
	err2 := New(http.StatusNotFound, "NOT_FOUND", "task not found")
	err3 := New(http.StatusNotFound, "NOT_FOUND", "user not found")
	err4 := New(http.StatusBadRequest, "BAD_REQUEST", "task not found")

	if !err1.Is(err2) {
		t.Fatal("expected err1.Is(err2) to be true")
	}
	if err1.Is(err3) {
		t.Fatal("expected err1.Is(err3) to be false (different message)")
	}
	if err1.Is(err4) {
		t.Fatal("expected err1.Is(err4) to be false (different status/code)")
	}
	if err1.Is(errors.New("standard error")) {
		t.Fatal("expected err1.Is to return false for standard error")
	}
}

func TestAppError_Wrap(t *testing.T) {
	template := New(http.StatusBadRequest, "BAD_REQUEST", "bad request")

	// Nil receiver
	var nilErr *AppError
	if nilErr.Wrap(errors.New("err")) != nil {
		t.Fatal("expected nil on nil receiver")
	}

	// Nil raw error
	if template.Wrap(nil) != template {
		t.Fatal("expected original template on nil rawErr")
	}

	// Raw error wrapping
	raw := errors.New("field title is required")
	wrapped := template.Wrap(raw)
	if wrapped.HTTPStatus != http.StatusBadRequest || wrapped.Code != "BAD_REQUEST" || wrapped.Message != "field title is required" {
		t.Fatalf("unexpected wrapped error: %+v", wrapped)
	}
	if !errors.Is(wrapped, raw) {
		t.Fatal("expected errors.Is to match wrapped raw error")
	}

	// Already an AppError
	existing := New(http.StatusNotFound, "NOT_FOUND", "not found")
	wrappedExisting := template.Wrap(existing)
	if wrappedExisting != existing {
		t.Fatal("expected Wrap to return existing AppError without double-wrapping")
	}

	// 5xx error preserves safe message
	serverTemplate := New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	dbErr := errors.New("pq: password authentication failed")
	wrappedServer := serverTemplate.Wrap(dbErr)
	if wrappedServer.Message != "internal server error" {
		t.Fatalf("expected 500 error to preserve safe message, got: %s", wrappedServer.Message)
	}
	if !errors.Is(wrappedServer, dbErr) {
		t.Fatal("expected 500 error to still wrap dbErr as RawErr")
	}
}

func TestAppError_WithMessage(t *testing.T) {
	template := New(http.StatusBadRequest, "BAD_REQUEST", "bad request")

	var nilErr *AppError
	if nilErr.WithMessage("msg") != nil {
		t.Fatal("expected nil on nil receiver")
	}

	updated := template.WithMessage("new custom message")
	if updated.Message != "new custom message" || updated.HTTPStatus != http.StatusBadRequest {
		t.Fatalf("unexpected updated error: %+v", updated)
	}
}

func TestAppError_StackTrace(t *testing.T) {
	// 5xx wrap captures a stack with a resolvable frame pointing at this test.
	serverTemplate := New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	wrapped := serverTemplate.Wrap(errors.New("boom"))

	frames := wrapped.StackTrace()
	if len(frames) == 0 {
		t.Fatal("expected 5xx wrap to capture a stack trace")
	}
	formatted := wrapped.FormattedStack()
	if formatted == "" {
		t.Fatal("expected non-empty formatted stack for 5xx error")
	}
	if !strings.Contains(formatted, "TestAppError_StackTrace") {
		t.Fatalf("expected formatted stack to reference the caller, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "apperror_test.go") {
		t.Fatalf("expected formatted stack to include file name, got:\n%s", formatted)
	}

	// 4xx wrap does not capture a stack (client errors are expected, not bugs).
	clientTemplate := New(http.StatusBadRequest, "BAD_REQUEST", "bad request")
	clientWrapped := clientTemplate.Wrap(errors.New("invalid field"))
	if len(clientWrapped.StackTrace()) != 0 || clientWrapped.FormattedStack() != "" {
		t.Fatal("expected 4xx wrap to capture no stack trace")
	}

	// Wrapping an existing AppError does not recapture the stack.
	existing := serverTemplate.Wrap(errors.New("first cause"))
	rewrapped := New(http.StatusInternalServerError, "INTERNAL_ERROR", "x").Wrap(existing)
	if rewrapped != existing {
		t.Fatal("expected Wrap to return the existing AppError unchanged")
	}

	// WithMessage preserves the captured stack.
	withMsg := wrapped.WithMessage("new message")
	if len(withMsg.StackTrace()) == 0 {
		t.Fatal("expected WithMessage to preserve the captured stack")
	}

	// Nil receiver and stack-less errors are safe.
	var nilErr *AppError
	if nilErr.StackTrace() != nil || nilErr.FormattedStack() != "" {
		t.Fatal("expected nil receiver to have no stack")
	}
	plain := New(http.StatusInternalServerError, "INTERNAL_ERROR", "no cause")
	if plain.FormattedStack() != "" {
		t.Fatal("expected New without a cause to capture no stack")
	}
}


