package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/pkg/ctxmeta"
)

// AuthValidator defines the required interface for authenticating incoming requests.
type AuthValidator interface {
	Authenticate(ctx context.Context, tokenStr string) (*ctxmeta.AuthUser, error)
}

// Auth returns the standard authentication middleware from the Middlewares container.
func (m *Middlewares) Auth() gin.HandlerFunc {
	return Auth(m.validator)
}

// OptionalAuth returns the optional authentication middleware from the Middlewares container.
func (m *Middlewares) OptionalAuth() gin.HandlerFunc {
	return OptionalAuth(m.validator)
}

// RequirePermission returns the permission check middleware from the Middlewares container.
func (m *Middlewares) RequirePermission(requiredPerm string) gin.HandlerFunc {
	return RequirePermission(requiredPerm)
}

// Auth returns a Gin middleware that validates the JWT Bearer token using the provided AuthValidator service.
func Auth(validator AuthValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, err.Error())
			return
		}

		if validator == nil {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, "invalid or expired token")
			return
		}

		authUser, err := validator.Authenticate(c.Request.Context(), tokenStr)
		if err != nil || authUser == nil {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, "invalid or expired token")
			return
		}

		// Inject authenticated user into standard context
		c.Request = c.Request.WithContext(ctxmeta.WithAuthUser(c.Request.Context(), *authUser))

		c.Next()
	}
}

// OptionalAuth returns a Gin middleware that validates the token if present,
// injecting the AuthUser into context, but does not reject unauthenticated requests.
func OptionalAuth(validator AuthValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || validator == nil {
			c.Next()
			return
		}

		tokenStr, err := extractBearerToken(authHeader)
		if err != nil {
			c.Next()
			return
		}

		authUser, err := validator.Authenticate(c.Request.Context(), tokenStr)
		if err == nil && authUser != nil {
			c.Request = c.Request.WithContext(ctxmeta.WithAuthUser(c.Request.Context(), *authUser))
		}

		c.Next()
	}
}

// RequirePermission checks whether the authenticated user has the required permission.
func RequirePermission(requiredPerm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authUser, ok := ctxmeta.GetAuthUser(c.Request.Context())
		if !ok || authUser.UserID == uuid.Nil {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, "unauthorized: authentication required")
			return
		}

		if !authUser.HasPermission(requiredPerm) {
			abortError(c, http.StatusForbidden, constants.ResponseCodeForbidden, "forbidden: insufficient permissions")
			return
		}

		c.Next()
	}
}
