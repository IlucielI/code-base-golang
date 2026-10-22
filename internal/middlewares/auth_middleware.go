package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"code-base-golang/internal/config"
	"code-base-golang/internal/constants"
	"code-base-golang/internal/pkg/ctxmeta"
	"code-base-golang/internal/pkg/jwt"
)

// Auth validates the JWT Bearer token using jwt.ValidateToken.
func (m *Middlewares) Auth(c *gin.Context) {
	Auth(m.cfg)(c)
}

// AnonAuth validates an anonymous guest token.
func (m *Middlewares) AnonAuth(c *gin.Context) {
	AnonAuth(m.cfg)(c)
}

// OptionalAuth validates the token if present, injecting the AuthUser into context,
// but does not reject unauthenticated requests.
func (m *Middlewares) OptionalAuth(c *gin.Context) {
	OptionalAuth(m.cfg)(c)
}

// RequirePermission returns the permission check middleware from the Middlewares container.
func (m *Middlewares) RequirePermission(requiredPerm string) gin.HandlerFunc {
	return RequirePermission(requiredPerm)
}

// Auth returns a Gin middleware that validates the JWT Bearer token against configuration.
func Auth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, err.Error())
			return
		}

		claims, err := jwt.ValidateToken(cfg, tokenStr)
		if err != nil || claims == nil {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, "invalid or expired token")
			return
		}

		role := strings.TrimSpace(claims.Role)
		if role == "" {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, "invalid token: missing role claim")
			return
		}

		authUser := ctxmeta.AuthUser{
			UserID:    claims.UserID,
			Email:     claims.Email,
			SessionID: claims.SessionID,
			Role:      role,
		}

		c.Request = c.Request.WithContext(ctxmeta.WithAuthUser(c.Request.Context(), authUser))
		c.Next()
	}
}

// AnonAuth returns a Gin middleware that validates an anonymous guest token.
func AnonAuth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, err.Error())
			return
		}

		claims, err := jwt.ValidateToken(cfg, tokenStr)
		if err != nil || claims == nil {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, "invalid or expired anonymous token")
			return
		}

		if claims.TokenType != "anon" && claims.Role != "anonymous" {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, "invalid anonymous token")
			return
		}

		authUser := ctxmeta.AuthUser{
			UserID: claims.UserID,
			Role:   "anonymous",
		}

		c.Request = c.Request.WithContext(ctxmeta.WithAuthUser(c.Request.Context(), authUser))
		c.Next()
	}
}

// OptionalAuth returns a Gin middleware that validates the token if present,
// injecting the AuthUser into context, but does not reject unauthenticated requests.
func OptionalAuth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenStr, err := extractBearerToken(authHeader)
		if err != nil {
			c.Next()
			return
		}

		claims, err := jwt.ValidateToken(cfg, tokenStr)
		if err == nil && claims != nil {
			role := strings.TrimSpace(claims.Role)
			if role != "" {
				authUser := ctxmeta.AuthUser{
					UserID:    claims.UserID,
					Email:     claims.Email,
					SessionID: claims.SessionID,
					Role:      role,
				}
				c.Request = c.Request.WithContext(ctxmeta.WithAuthUser(c.Request.Context(), authUser))
			}
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
