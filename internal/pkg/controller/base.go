package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"code-base-golang/internal/config"
	"code-base-golang/internal/constants"
	"code-base-golang/internal/dtos"
)

// Base provides shared response formatting and error handling helpers for all controllers.
type Base struct {
	Cfg       config.Config
	StartedAt time.Time
}

// NewBase constructs a new Base controller instance.
func NewBase(cfg config.Config) Base {
	return Base{
		Cfg:       cfg,
		StartedAt: time.Now(),
	}
}

// WrapError translates any error into a standard JSON response using AppError metadata.
func (b *Base) WrapError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

	appErr := constants.ErrInternalServerError.Wrap(err)

	// Record the error on the active trace span (no-op when tracing is disabled).
	if ctx.Request != nil {
		if span := trace.SpanFromContext(ctx.Request.Context()); span.IsRecording() {
			span.RecordError(err)
			span.SetStatus(codes.Error, appErr.Message)
			if stackTrace := appErr.FormattedStack(); stackTrace != "" {
				span.SetAttributes(attribute.String("exception.stacktrace", stackTrace))
			}
		}
	}

	status := constants.ResponseStatusError
	if appErr.HTTPStatus >= 400 && appErr.HTTPStatus < 500 {
		status = constants.ResponseStatusFail
	} else if appErr.HTTPStatus >= 500 {
		reqMethod := "UNKNOWN"
		reqPath := "UNKNOWN"
		if ctx.Request != nil {
			reqMethod = ctx.Request.Method
			if ctx.Request.URL != nil {
				reqPath = ctx.Request.URL.Path
			}
		}
		log.Printf("[API ERROR %d] %s %s: %v", appErr.HTTPStatus, reqMethod, reqPath, err)
		// In development, print the captured stack trace to aid local debugging.
		if b.Cfg.AppEnv == "development" {
			if stackTrace := appErr.FormattedStack(); stackTrace != "" {
				log.Printf("[API ERROR %d] stack trace:\n%s", appErr.HTTPStatus, stackTrace)
			}
		}
	}

	ctx.JSON(appErr.HTTPStatus, dtos.BaseResponse{
		Status:    status,
		Code:      appErr.Code,
		Message:   appErr.Message,
		Timestamp: time.Now(),
	})
}

// RespondSuccess sends a standardized JSON success response envelope.
func (b *Base) RespondSuccess(ctx *gin.Context, httpStatus int, message string, data any) {
	ctx.JSON(httpStatus, dtos.APIResponse[any]{
		Status:    constants.ResponseStatusSuccess,
		Code:      constants.ResponseCodeSuccess,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// RespondOK sends a standard HTTP 200 OK JSON success envelope.
func (b *Base) RespondOK(ctx *gin.Context, message string, data any) {
	b.RespondSuccess(ctx, http.StatusOK, message, data)
}

// RespondCreated sends a standard HTTP 201 Created JSON success envelope.
func (b *Base) RespondCreated(ctx *gin.Context, message string, data any) {
	b.RespondSuccess(ctx, http.StatusCreated, message, data)
}

// RespondEmpty sends a standard HTTP 200 OK JSON response without a data payload.
func (b *Base) RespondEmpty(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusOK, dtos.BaseResponse{
		Status:    constants.ResponseStatusSuccess,
		Code:      constants.ResponseCodeSuccess,
		Message:   message,
		Timestamp: time.Now(),
	})
}
