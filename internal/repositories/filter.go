package repositories

import (
	"fmt"
	"strings"
)

// PaginationFilter encapsulates database pagination parameters.
type PaginationFilter struct {
	Limit  int
	Offset int
}

// NewPaginationFilter constructs a PaginationFilter converting 1-based page to 0-based database offset.
// If limit <= 0, it falls back to defaultLimit (or 10 if defaultLimit <= 0).
func NewPaginationFilter(page, limit int, defaultLimit ...int) PaginationFilter {
	if limit <= 0 {
		if len(defaultLimit) > 0 && defaultLimit[0] > 0 {
			limit = defaultLimit[0]
		} else {
			limit = 10
		}
	}
	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}
	return PaginationFilter{
		Limit:  limit,
		Offset: offset,
	}
}

// SortFilter encapsulates database sorting parameters.
type SortFilter struct {
	SortBy string // e.g. "created_at"
	Order  string // "asc" or "desc"
}

// OrderClause returns a sanitized, formatted SQL ORDER BY clause (e.g. "created_at DESC").
// If SortBy is empty and a fallback is provided, the fallback is returned directly.
func (s SortFilter) OrderClause(fallback ...string) string {
	sortBy := strings.TrimSpace(s.SortBy)
	if sortBy == "" {
		if len(fallback) > 0 {
			return strings.TrimSpace(fallback[0])
		}
		return ""
	}
	order := "ASC"
	if strings.EqualFold(strings.TrimSpace(s.Order), "desc") {
		order = "DESC"
	}
	return fmt.Sprintf("%s %s", sortBy, order)
}
