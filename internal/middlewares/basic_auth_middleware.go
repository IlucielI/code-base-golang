package middlewares

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/constants"
)

// BasicAuth returns the HTTP Basic Authentication middleware from the Middlewares container.
// It defaults to the configured BasicAuthUsername and BasicAuthPassword, or accepts explicit credentials.
func (m *Middlewares) BasicAuth(userAndPass ...string) gin.HandlerFunc {
	user := m.cfg.BasicAuthUsername
	pass := m.cfg.BasicAuthPassword
	if len(userAndPass) >= 2 {
		user = userAndPass[0]
		pass = userAndPass[1]
	}
	return BasicAuth(user, pass)
}

// BasicAuth returns a Gin middleware that validates HTTP Basic Authentication credentials.
// It verifies the Authorization header format ('Basic <base64>') and matches the credentials
// in constant time against expectedUser and expectedPass.
func BasicAuth(expectedUser, expectedPass string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Basic") {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, "invalid authorization header format, expected 'Basic <credentials>'")
			return
		}

		payload, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, "invalid basic auth encoding")
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, "invalid basic auth credentials format, expected 'username:password'")
			return
		}

		userMatch := subtle.ConstantTimeCompare([]byte(pair[0]), []byte(expectedUser)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(pair[1]), []byte(expectedPass)) == 1

		if !userMatch || !passMatch {
			abortError(c, http.StatusUnauthorized, constants.ResponseCodeUnauthorized, "invalid basic auth credentials")
			return
		}

		c.Next()
	}
}
