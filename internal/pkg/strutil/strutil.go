// Package strutil provides small, dependency-light string helpers that are
// reusable across layers without any domain coupling.
package strutil

import (
	"crypto/subtle"
	"strings"
)

// ConstantTimeEqual performs a constant-time comparison of two strings to prevent
// timing side channels (e.g. when comparing tokens, API keys, or signatures).
func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// LikePattern returns a sanitized SQL LIKE wildcard pattern "%<trimmed-query>%".
// If the input is empty or contains only whitespace, it returns an empty string
// so callers can skip the filter entirely.
func LikePattern(query string) string {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return ""
	}
	return "%" + trimmed + "%"
}
