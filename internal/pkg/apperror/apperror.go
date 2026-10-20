package apperror

import (
	"errors"
	"runtime"
	"strconv"
	"strings"
)

// AppError represents an application-level domain error carrying HTTP status and response codes.
type AppError struct {
	HTTPStatus int
	Code       string
	Message    string
	RawErr     error
	// stack holds program counters captured at the point a non-AppError was first
	// wrapped into a server-side (5xx) AppError. It is nil for client (4xx) errors
	// and for templates created via New without a cause.
	stack []uintptr
}

// stackDepth bounds how many frames are captured per error.
const stackDepth = 32

// captureStack records the current call stack, skipping the given number of
// internal frames so the first recorded frame is the caller of interest.
func captureStack(skip int) []uintptr {
	pcs := make([]uintptr, stackDepth)
	n := runtime.Callers(skip, pcs)
	return pcs[:n]
}

// StackTrace returns the captured stack as a slice of resolved frames. It is
// empty when no stack was captured (client errors or errors without a cause).
func (e *AppError) StackTrace() []runtime.Frame {
	if e == nil || len(e.stack) == 0 {
		return nil
	}
	frames := runtime.CallersFrames(e.stack)
	out := make([]runtime.Frame, 0, len(e.stack))
	for {
		frame, more := frames.Next()
		out = append(out, frame)
		if !more {
			break
		}
	}
	return out
}

// FormattedStack renders the captured stack as newline-separated
// "func\n\tfile:line" entries, or an empty string when no stack is present.
func (e *AppError) FormattedStack() string {
	frames := e.StackTrace()
	if len(frames) == 0 {
		return ""
	}
	var b strings.Builder
	for _, f := range frames {
		b.WriteString(f.Function)
		b.WriteString("\n\t")
		b.WriteString(f.File)
		b.WriteString(":")
		b.WriteString(strconv.Itoa(f.Line))
		b.WriteString("\n")
	}
	return b.String()
}

// Error satisfies the Go standard error interface.
func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.RawErr != nil {
		return e.RawErr.Error()
	}
	return ""
}

// Unwrap returns the underlying raw error for errors.Is / errors.As chaining.
func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.RawErr
}

// Is compares equality with another error or AppError.
func (e *AppError) Is(target error) bool {
	if e == nil || target == nil {
		return e == nil && target == nil
	}
	var targetAppErr *AppError
	if errors.As(target, &targetAppErr) {
		return e.HTTPStatus == targetAppErr.HTTPStatus &&
			e.Code == targetAppErr.Code &&
			e.Message == targetAppErr.Message
	}
	return false
}

// New constructs a new AppError.
func New(httpStatus int, code string, message string, rawErr ...error) *AppError {
	var err error
	if len(rawErr) > 0 {
		err = rawErr[0]
	}
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
		RawErr:     err,
	}
}

// Wrap returns a copy of AppError wrapping rawErr.
// If rawErr is already an *AppError, it is returned unchanged.
// For client errors (HTTPStatus < 500), rawErr.Error() is used as Message to expose validation details.
// For server errors (HTTPStatus >= 500), the template's safe Message is preserved to prevent internal data leaks.
func (e *AppError) Wrap(rawErr error) *AppError {
	if e == nil {
		return nil
	}
	if rawErr == nil {
		return e
	}
	var existing *AppError
	if errors.As(rawErr, &existing) {
		return existing
	}

	msg := e.Message
	var stack []uintptr
	if e.HTTPStatus < 500 {
		// Client errors expose the cause's message and need no stack trace.
		msg = rawErr.Error()
	} else {
		// Server errors capture the stack at this first wrap point (closest to
		// the cause) for diagnostics. skip=3: runtime.Callers, captureStack, Wrap.
		stack = captureStack(3)
	}

	return &AppError{
		HTTPStatus: e.HTTPStatus,
		Code:       e.Code,
		Message:    msg,
		RawErr:     rawErr,
		stack:      stack,
	}
}

// WithMessage returns a copy of AppError with an updated message, preserving any
// captured stack trace.
func (e *AppError) WithMessage(msg string) *AppError {
	if e == nil {
		return nil
	}
	return &AppError{
		HTTPStatus: e.HTTPStatus,
		Code:       e.Code,
		Message:    msg,
		RawErr:     e.RawErr,
		stack:      e.stack,
	}
}


