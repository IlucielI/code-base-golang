package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/pkg/apperror"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestResponse_OK(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	OK(ctx, "success message", map[string]string{"foo": "bar"})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp APIResponse[map[string]string]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if resp.Status != constants.ResponseStatusSuccess {
		t.Errorf("expected success status, got %s", resp.Status)
	}
	if resp.Data["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %s", resp.Data["foo"])
	}
}

func TestResponse_Created(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	Created(ctx, "created message", map[string]string{"id": "123"})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestResponse_Empty(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	Empty(ctx, "empty message")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestResponse_Error(t *testing.T) {
	t.Run("nil error does nothing", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		Error(ctx, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on nil error, got %d", w.Code)
		}
	})

	t.Run("AppError renders exact HTTP status and code", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		appErr := apperror.New(http.StatusConflict, "CONFLICT", "resource conflict")
		Error(ctx, appErr)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", w.Code)
		}
		var resp BaseResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Status != constants.ResponseStatusFail || resp.Code != "CONFLICT" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("raw unclassified error falls back to 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		Error(ctx, errors.New("unhandled failure"))

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}
