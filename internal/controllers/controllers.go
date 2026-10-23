package controllers

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
	"code-base-golang/internal/usecases"
)

// Controllers is the central container for all handler methods.
type Controllers struct {
	cfg       config.Config
	uc        *usecases.Usecases
	startedAt time.Time
}

// New initializes the central controllers container with application dependencies.
func New(cfg config.Config, uc *usecases.Usecases) *Controllers {
	return &Controllers{
		cfg:       cfg,
		uc:        uc,
		startedAt: time.Now(),
	}
}

// SetUsecases allows overriding or injecting custom / mock Usecases for tests.
func (c *Controllers) SetUsecases(u *usecases.Usecases) {
	c.uc = u
}

// Usecases returns the underlying usecases container.
func (c *Controllers) Usecases() *usecases.Usecases {
	return c.uc
}

// wrapError translates any error into a standard JSON response using AppError metadata.
// Server-side (5xx) errors are logged with request method and path for observability,
// while the client only ever receives the sanitized AppError message.
func (c *Controllers) wrapError(ctx *gin.Context, err error) {
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
		// In development, print the capture stack trace to aid local debugging.
		if c.cfg.AppEnv == "development" {
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

// respondSuccess sends a standardized JSON success response envelope.
func (c *Controllers) respondSuccess(ctx *gin.Context, httpStatus int, message string, data any) {
	ctx.JSON(httpStatus, dtos.APIResponse[any]{
		Status:    constants.ResponseStatusSuccess,
		Code:      constants.ResponseCodeSuccess,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// respondOK sends a standard HTTP 200 OK JSON success envelope.
func (c *Controllers) respondOK(ctx *gin.Context, message string, data any) {
	c.respondSuccess(ctx, http.StatusOK, message, data)
}

// respondCreated sends a standard HTTP 201 Created JSON success envelope.
func (c *Controllers) respondCreated(ctx *gin.Context, message string, data any) {
	c.respondSuccess(ctx, http.StatusCreated, message, data)
}

// respondEmpty sends a standard HTTP 200 OK JSON response without a data payload.
func (c *Controllers) respondEmpty(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusOK, dtos.BaseResponse{
		Status:    constants.ResponseStatusSuccess,
		Code:      constants.ResponseCodeSuccess,
		Message:   message,
		Timestamp: time.Now(),
	})
}
