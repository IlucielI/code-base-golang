package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwtgo "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"code-base-golang/internal/config"
	"code-base-golang/internal/pkg/ctxmeta"
	"code-base-golang/internal/pkg/jwt"
)

func setupTestAuthConfig() config.Config {
	return config.Config{
		AppName:             "test-app",
		JWTSecret:           "auth-test-secret-key-that-is-long-enough",
		JWTAccessExpiration: 15 * time.Minute,
	}
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := setupTestAuthConfig()
	mws := New(cfg)

	userID := uuid.New()
	email := "user@example.com"
	tokens, err := jwt.GenerateTokenPair(cfg, userID, email, "session-123")
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	// 1. Missing Authorization header -> 401
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(mws.Auth)
		r.GET("/protected", func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/protected", nil)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 on missing auth header, got %d", w.Code)
		}
	}

	// 2. Invalid Authorization header format -> 401
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(mws.Auth)
		r.GET("/protected", func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/protected", nil)
		c.Request.Header.Set("Authorization", "Basic 123456")
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 on invalid format, got %d", w.Code)
		}
	}

	// 3. Invalid JWT token signature/malformed -> 401
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(mws.Auth)
		r.GET("/protected", func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/protected", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid.token.value")
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 on invalid token, got %d", w.Code)
		}
	}

	// 4. Valid JWT token -> 200 & injects AuthUser
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		var capturedUser ctxmeta.AuthUser
		r.Use(mws.Auth)
		r.GET("/protected", func(ctx *gin.Context) {
			u, ok := ctxmeta.GetAuthUser(ctx.Request.Context())
			if ok {
				capturedUser = u
			}
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/protected", nil)
		c.Request.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on valid token, got %d", w.Code)
		}
		if capturedUser.UserID != userID || capturedUser.Email != email || capturedUser.SessionID != "session-123" || capturedUser.Role != "user" {
			t.Errorf("unexpected captured user: %+v", capturedUser)
		}
	}

	// 5. Valid JWT token with custom role -> 200 & captures custom role
	{
		adminTokens, err := jwt.GenerateTokenPair(cfg, userID, "admin@example.com", "session-admin", "admin")
		if err != nil {
			t.Fatalf("failed to generate admin token: %v", err)
		}

		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		var capturedUser ctxmeta.AuthUser
		r.Use(mws.Auth)
		r.GET("/protected-role", func(ctx *gin.Context) {
			u, ok := ctxmeta.GetAuthUser(ctx.Request.Context())
			if ok {
				capturedUser = u
			}
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/protected-role", nil)
		c.Request.Header.Set("Authorization", "Bearer "+adminTokens.AccessToken)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on valid admin token, got %d", w.Code)
		}
		if capturedUser.Role != "admin" {
			t.Errorf("expected role 'admin', got '%s'", capturedUser.Role)
		}
	}

	// 6. Token with missing role claim -> 401
	{
		now := time.Now()
		noRoleClaims := jwt.CustomClaims{
			UserID:    userID,
			SessionID: "sess-no-role",
			Email:     "norole@example.com",
			Role:      "",
			TokenType: "access",
			RegisteredClaims: jwtgo.RegisteredClaims{
				Subject:   userID.String(),
				Issuer:    cfg.AppName,
				IssuedAt:  jwtgo.NewNumericDate(now),
				ExpiresAt: jwtgo.NewNumericDate(now.Add(cfg.JWTAccessExpiration)),
			},
		}
		tokenObj := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, noRoleClaims)
		tokenStr, err := tokenObj.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			t.Fatalf("failed to sign token without role: %v", err)
		}

		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(mws.Auth)
		r.GET("/protected", func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/protected", nil)
		c.Request.Header.Set("Authorization", "Bearer "+tokenStr)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 on token with missing role claim, got %d", w.Code)
		}
	}
}

func TestOptionalAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := setupTestAuthConfig()
	mws := New(cfg)

	userID := uuid.New()
	email := "opt@example.com"
	tokens, err := jwt.GenerateTokenPair(cfg, userID, email, "session-opt")
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	// 1. Without Authorization header -> 200, no user injected
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(mws.OptionalAuth)
		r.GET("/optional", func(ctx *gin.Context) {
			_, ok := ctxmeta.GetAuthUser(ctx.Request.Context())
			if ok {
				t.Error("expected no AuthUser in context when header missing")
			}
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/optional", nil)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on missing header in OptionalAuth, got %d", w.Code)
		}
	}

	// 2. With valid Authorization header -> 200, user injected
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		var capturedUser ctxmeta.AuthUser
		r.Use(mws.OptionalAuth)
		r.GET("/optional", func(ctx *gin.Context) {
			u, ok := ctxmeta.GetAuthUser(ctx.Request.Context())
			if ok {
				capturedUser = u
			}
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/optional", nil)
		c.Request.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if capturedUser.UserID != userID || capturedUser.Email != email {
			t.Errorf("unexpected captured user in OptionalAuth: %+v", capturedUser)
		}
	}

	// 3. Token without role claim -> 200 but does NOT inject AuthUser
	{
		now := time.Now()
		noRoleClaims := jwt.CustomClaims{
			UserID:    userID,
			SessionID: "sess-opt-no-role",
			Email:     "norole-opt@example.com",
			Role:      "",
			TokenType: "access",
			RegisteredClaims: jwtgo.RegisteredClaims{
				Subject:   userID.String(),
				Issuer:    cfg.AppName,
				IssuedAt:  jwtgo.NewNumericDate(now),
				ExpiresAt: jwtgo.NewNumericDate(now.Add(cfg.JWTAccessExpiration)),
			},
		}
		tokenObj := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, noRoleClaims)
		tokenStr, err := tokenObj.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			t.Fatalf("failed to sign token without role: %v", err)
		}

		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		injected := false
		r.Use(mws.OptionalAuth)
		r.GET("/optional", func(ctx *gin.Context) {
			_, ok := ctxmeta.GetAuthUser(ctx.Request.Context())
			injected = ok
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/optional", nil)
		c.Request.Header.Set("Authorization", "Bearer "+tokenStr)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on OptionalAuth even with malformed role, got %d", w.Code)
		}
		if injected {
			t.Error("expected AuthUser NOT to be injected when role claim is missing")
		}
	}
}

func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := setupTestAuthConfig()
	mws := New(cfg)

	// 1. Unauthenticated -> 401
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(mws.RequirePermission("users:read"))
		r.GET("/perm-test", func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/perm-test", nil)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 on missing auth context, got %d", w.Code)
		}
	}

	// 2. Insufficient permission -> 403
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(func(ctx *gin.Context) {
			ctx.Request = ctx.Request.WithContext(ctxmeta.WithAuthUser(ctx.Request.Context(), ctxmeta.AuthUser{
				UserID:      uuid.New(),
				Email:       "user@example.com",
				Role:        "guest",
				Permissions: []string{"users:read"},
			}))
			ctx.Next()
		})
		r.Use(mws.RequirePermission("users:delete"))
		r.GET("/perm-test", func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/perm-test", nil)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 on insufficient permission, got %d", w.Code)
		}
	}

	// 3. Matching permission -> 200
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(func(ctx *gin.Context) {
			ctx.Request = ctx.Request.WithContext(ctxmeta.WithAuthUser(ctx.Request.Context(), ctxmeta.AuthUser{
				UserID:      uuid.New(),
				Email:       "user@example.com",
				Role:        "member",
				Permissions: []string{"users:read"},
			}))
			ctx.Next()
		})
		r.Use(mws.RequirePermission("users:read"))
		r.GET("/perm-test", func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/perm-test", nil)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on matching permission, got %d", w.Code)
		}
	}

	// 4. Wildcard permission "*" -> 200
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(func(ctx *gin.Context) {
			ctx.Request = ctx.Request.WithContext(ctxmeta.WithAuthUser(ctx.Request.Context(), ctxmeta.AuthUser{
				UserID:      uuid.New(),
				Email:       "admin@example.com",
				Role:        "admin",
				Permissions: []string{"*"},
			}))
			ctx.Next()
		})
		r.Use(mws.RequirePermission("users:delete"))
		r.GET("/perm-test", func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/perm-test", nil)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on wildcard permission, got %d", w.Code)
		}
	}
}

func TestAnonAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := setupTestAuthConfig()
	mws := New(cfg)

	anonToken, err := jwt.GenerateAnonToken(cfg)
	if err != nil {
		t.Fatalf("failed to generate anon token: %v", err)
	}

	userToken, err := jwt.GenerateTokenPair(cfg, uuid.New(), "user@example.com", "sess-1")
	if err != nil {
		t.Fatalf("failed to generate user token: %v", err)
	}

	// 1. Missing header -> 401
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(mws.AnonAuth)
		r.GET("/anon-only", func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/anon-only", nil)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 on missing anon auth header, got %d", w.Code)
		}
	}

	// 2. Non-anon token provided -> 401
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(mws.AnonAuth)
		r.GET("/anon-only", func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/anon-only", nil)
		c.Request.Header.Set("Authorization", "Bearer "+userToken.AccessToken)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 when passing standard user token to AnonAuth, got %d", w.Code)
		}
	}

	// 3. Valid anon token -> 200 & injects AuthUser with Role "anonymous"
	{
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		var capturedUser ctxmeta.AuthUser
		r.Use(mws.AnonAuth)
		r.GET("/anon-only", func(ctx *gin.Context) {
			u, ok := ctxmeta.GetAuthUser(ctx.Request.Context())
			if ok {
				capturedUser = u
			}
			ctx.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodGet, "/anon-only", nil)
		c.Request.Header.Set("Authorization", "Bearer "+anonToken.AccessToken)
		r.ServeHTTP(w, c.Request)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on valid anon token, got %d", w.Code)
		}
		if capturedUser.Role != "anonymous" {
			t.Errorf("expected role 'anonymous', got '%s'", capturedUser.Role)
		}
		if capturedUser.UserID == uuid.Nil {
			t.Error("expected valid anon UserID, got Nil")
		}
	}
}
