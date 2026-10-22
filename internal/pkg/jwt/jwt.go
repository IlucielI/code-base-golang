package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"code-base-golang/internal/config"
	"code-base-golang/internal/constants"
	"code-base-golang/internal/dtos"
)

// CustomClaims encapsulates application-specific payload fields in standard JWT claims.
type CustomClaims struct {
	UserID    uuid.UUID `json:"sub"`
	SessionID string    `json:"sid"`
	Email     string    `json:"email"`
	Role      string    `json:"role,omitempty"`
	TokenType string    `json:"type"` // "access", "refresh", or "anon"
	jwt.RegisteredClaims
}

// GenerateTokenPair signs and returns a new Access Token and Refresh Token pair.
// An optional role can be specified (defaults to "user").
func GenerateTokenPair(cfg config.Config, userID uuid.UUID, email string, sessionID string, role ...string) (*dtos.TokenResponse, error) {
	now := time.Now()
	accessExp := now.Add(cfg.JWTAccessExpiration)
	refreshExp := now.Add(cfg.JWTRefreshExpiration)

	assignedRole := constants.RoleUser
	if len(role) > 0 && role[0] != "" {
		assignedRole = role[0]
	}

	// 1. Access Token
	accessClaims := CustomClaims{
		UserID:    userID,
		SessionID: sessionID,
		Email:     email,
		Role:      assignedRole,
		TokenType: constants.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    cfg.AppName,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExp),
		},
	}
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := accessTokenObj.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// 2. Refresh Token
	refreshClaims := CustomClaims{
		UserID:    userID,
		SessionID: sessionID,
		Email:     email,
		TokenType: constants.TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    cfg.AppName,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExp),
		},
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &dtos.TokenResponse{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenType:        constants.TokenTypeBearer,
		ExpiresIn:        int64(cfg.JWTAccessExpiration.Seconds()),
		RefreshExpiresIn: int64(cfg.JWTRefreshExpiration.Seconds()),
	}, nil
}

// ValidateToken parses and verifies the signature and expiration of a JWT token string.
func ValidateToken(cfg config.Config, tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, constants.ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, constants.ErrInvalidToken
	}

	return claims, nil
}

// GenerateAnonToken signs and returns an anonymous guest token.
func GenerateAnonToken(cfg config.Config) (*dtos.TokenResponse, error) {
	now := time.Now()
	exp := now.Add(cfg.JWTAccessExpiration)
	guestID := uuid.New()

	claims := CustomClaims{
		UserID:    guestID,
		Role:      constants.RoleAnonymous,
		TokenType: constants.TokenTypeAnon,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   guestID.String(),
			Issuer:    cfg.AppName,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := tokenObj.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign anon token: %w", err)
	}

	return &dtos.TokenResponse{
		AccessToken: tokenStr,
		TokenType:   constants.TokenTypeBearer,
		ExpiresIn:   int64(cfg.JWTAccessExpiration.Seconds()),
	}, nil
}
