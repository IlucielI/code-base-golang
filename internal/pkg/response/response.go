package response

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"code-base-golang/internal/constants"
)

// APIResponse represents the standard API envelope response.
type APIResponse[T any] struct {
	Status    string    `json:"status"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Data      T         `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// BaseResponse represents a response without data payload (e.g. errors or simple acknowledgments).
type BaseResponse struct {
	Status    string    `json:"status"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// OK sends a standard HTTP 200 OK JSON success envelope.
func OK(ctx *gin.Context, message string, data any) {
	ctx.JSON(http.StatusOK, APIResponse[any]{
		Status:    constants.ResponseStatusSuccess,
		Code:      constants.ResponseCodeSuccess,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// Created sends a standard HTTP 201 Created JSON success envelope.
func Created(ctx *gin.Context, message string, data any) {
	ctx.JSON(http.StatusCreated, APIResponse[any]{
		Status:    constants.ResponseStatusSuccess,
		Code:      constants.ResponseCodeSuccess,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// Empty sends a standard HTTP 200 OK JSON response without a data payload.
func Empty(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusOK, BaseResponse{
		Status:    constants.ResponseStatusSuccess,
		Code:      constants.ResponseCodeSuccess,
		Message:   message,
		Timestamp: time.Now(),
	})
}

// Error translates any error into a standard JSON response using AppError metadata.
func Error(ctx *gin.Context, err error) {
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
		if stackTrace := appErr.FormattedStack(); stackTrace != "" {
			log.Printf("[API ERROR %d] stack trace:\n%s", appErr.HTTPStatus, stackTrace)
		}
	}

	ctx.JSON(appErr.HTTPStatus, BaseResponse{
		Status:    status,
		Code:      appErr.Code,
		Message:   appErr.Message,
		Timestamp: time.Now(),
	})
}
