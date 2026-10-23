package middlewares_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/config"
	"code-base-golang/internal/constants"
	"code-base-golang/internal/middlewares"
	"code-base-golang/internal/pkg/ctxmeta"
	"code-base-golang/internal/pkg/response"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRecovery_NormalExecution(t *testing.T) {
	buf := &bytes.Buffer{}
	mws := middlewares.New(config.Config{})
	r := gin.New()
	r.Use(mws.Recovery(middlewares.WithRecoveryWriter(buf)))

	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/ok", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if buf.Len() > 0 {
		t.Errorf("expected no recovery log on normal execution, got: %s", buf.String())
	}
}

func TestRecovery_CatchesPanicAndReturnsUnifiedEnvelope(t *testing.T) {
	buf := &bytes.Buffer{}
	mws := middlewares.New(config.Config{})
	r := gin.New()
	r.Use(mws.Recovery(middlewares.WithRecoveryWriter(buf)))

	secretError := "database connection unexpectedly dropped: secret_token_xyz"
	r.GET("/panic", func(c *gin.Context) {
		panic(secretError)
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/panic", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 on panic recovery, got %d", w.Code)
	}

	var resp response.BaseResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response envelope: %v", err)
	}

	if resp.Status != constants.ResponseStatusError {
		t.Errorf("expected status 'error', got %s", resp.Status)
	}
	if resp.Code != constants.ResponseCodeInternalError {
		t.Errorf("expected code %s, got %s", constants.ResponseCodeInternalError, resp.Code)
	}
	if resp.Message != constants.ErrInternalServerError.Message {
		t.Errorf("expected message %s, got %s", constants.ErrInternalServerError.Message, resp.Message)
	}
	if resp.Timestamp.IsZero() {
		t.Errorf("expected non-zero timestamp, got %v", resp.Timestamp)
	}

	// Verify internal panic details are NOT exposed to client
	bodyStr := w.Body.String()
	if strings.Contains(bodyStr, "secret_token_xyz") {
		t.Errorf("expected secret error to not leak in response, but found in body: %s", bodyStr)
	}

	// Verify server-side logging captures panic message and stack
	logStr := buf.String()
	if !strings.Contains(logStr, "[PANIC RECOVERED]") {
		t.Errorf("expected log to contain '[PANIC RECOVERED]', got: %s", logStr)
	}
	if !strings.Contains(logStr, secretError) {
		t.Errorf("expected log to contain panic error message, got: %s", logStr)
	}
	if !strings.Contains(logStr, "stack:") {
		t.Errorf("expected log to contain stack trace, got: %s", logStr)
	}
	if !strings.Contains(logStr, "method=GET path=/panic") {
		t.Errorf("expected server-side log to contain method and path, got: %s", logStr)
	}
}

func TestRecovery_WithRequestIDInContext(t *testing.T) {
	buf := &bytes.Buffer{}
	mws := middlewares.New(config.Config{})
	r := gin.New()
	r.Use(mws.Recovery(middlewares.WithRecoveryWriter(buf)))

	r.GET("/panic-with-id", func(c *gin.Context) {
		panic("nil pointer dereference")
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/panic-with-id", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	testReqID := "custom-req-id-12345"
	ctx := ctxmeta.WithRequestID(req.Context(), testReqID)
	req = req.WithContext(ctx)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	logStr := buf.String()
	if !strings.Contains(logStr, "request_id="+testReqID) {
		t.Errorf("expected server log to include request_id=%s, got: %s", testReqID, logStr)
	}
}

func TestRecovery_WithRequestIDInHeaderOrContextSet(t *testing.T) {
	buf := &bytes.Buffer{}
	mws := middlewares.New(config.Config{})
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("request_id", "gin-context-req-id")
		c.Next()
	})
	r.Use(mws.Recovery(middlewares.WithRecoveryWriter(buf)))

	r.GET("/panic-header", func(c *gin.Context) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/panic-header", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("X-Request-ID", "header-req-id")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	logStr := buf.String()
	if !strings.Contains(logStr, "request_id=gin-context-req-id") {
		t.Errorf("expected server log to have request_id from gin context, got: %s", logStr)
	}
}
