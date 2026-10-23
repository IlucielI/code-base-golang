package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/config"
	"code-base-golang/internal/constants"
	"code-base-golang/internal/dtos"
	"code-base-golang/internal/pkg/apperror"
	"code-base-golang/internal/pkg/controller"
)

func TestBase_WrapError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	base := controller.NewBase(config.Config{})

	t.Run("nil error does nothing", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		base.WrapError(ctx, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected default 200 on nil error, got %d", w.Code)
		}
	})

	t.Run("AppError renders exact HTTP status and code", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		appErr := apperror.New(http.StatusConflict, "CONFLICT", "resource conflict")
		base.WrapError(ctx, appErr)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", w.Code)
		}
		var resp dtos.BaseResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Status != constants.ResponseStatusFail || resp.Code != "CONFLICT" || resp.Message != "resource conflict" {
			t.Fatalf("unexpected response payload: %+v", resp)
		}
	})

	t.Run("raw unclassified error falls back to ErrInternalServerError", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		base.WrapError(ctx, errors.New("unhandled failure"))

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
		var resp dtos.BaseResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Code != constants.ResponseCodeInternalError {
			t.Fatalf("expected code %s, got %s", constants.ResponseCodeInternalError, resp.Code)
		}
	})
}

func TestBase_RespondHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	base := controller.NewBase(config.Config{})

	t.Run("RespondOK sends 200 OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		base.RespondOK(ctx, "ok", map[string]string{"foo": "bar"})
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("RespondCreated sends 201 Created", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		base.RespondCreated(ctx, "created", map[string]string{"id": "1"})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", w.Code)
		}
	})

	t.Run("RespondEmpty sends 200 without data", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		base.RespondEmpty(ctx, "done")
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}
