package constants

// Token types for authentication and JWT claims.
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
	TokenTypeAnon    = "anon"
	TokenTypeBearer  = "Bearer"
)

// Standard user roles for authorization and access control.
const (
	RoleAdmin     = "admin"
	RoleUser      = "user"
	RoleAnonymous = "anonymous"
)
