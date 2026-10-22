package dtos

import (
	"testing"
)

func TestPaginationQuery_SetDefaults(t *testing.T) {
	tests := []struct {
		name          string
		initial       PaginationQuery
		customDefault []int
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "zero_values_use_defaults",
			initial:       PaginationQuery{Page: 0, Limit: 0},
			expectedPage:  1,
			expectedLimit: DefaultPaginationLimit,
		},
		{
			name:          "negative_values_normalized",
			initial:       PaginationQuery{Page: -5, Limit: -10},
			expectedPage:  1,
			expectedLimit: DefaultPaginationLimit,
		},
		{
			name:          "valid_values_preserved",
			initial:       PaginationQuery{Page: 2, Limit: 50},
			expectedPage:  2,
			expectedLimit: 50,
		},
		{
			name:          "limit_exceeding_max_is_capped",
			initial:       PaginationQuery{Page: 1, Limit: 250},
			expectedPage:  1,
			expectedLimit: MaxPaginationLimit,
		},
		{
			name:          "custom_default_limit_applied",
			initial:       PaginationQuery{Page: 1, Limit: 0},
			customDefault: []int{15},
			expectedPage:  1,
			expectedLimit: 15,
		},
		{
			name:          "invalid_custom_default_falls_back",
			initial:       PaginationQuery{Page: 1, Limit: 0},
			customDefault: []int{0},
			expectedPage:  1,
			expectedLimit: DefaultPaginationLimit,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q := tc.initial
			q.SetDefaults(tc.customDefault...)
			if q.Page != tc.expectedPage {
				t.Errorf("expected Page=%d, got %d", tc.expectedPage, q.Page)
			}
			if q.Limit != tc.expectedLimit {
				t.Errorf("expected Limit=%d, got %d", tc.expectedLimit, q.Limit)
			}
		})
	}
}

func TestPaginationQuery_Offset(t *testing.T) {
	tests := []struct {
		name     string
		query    PaginationQuery
		expected int
	}{
		{
			name:     "page_1",
			query:    PaginationQuery{Page: 1, Limit: 20},
			expected: 0,
		},
		{
			name:     "page_2",
			query:    PaginationQuery{Page: 2, Limit: 20},
			expected: 20,
		},
		{
			name:     "page_5",
			query:    PaginationQuery{Page: 5, Limit: 10},
			expected: 40,
		},
		{
			name:     "page_zero_or_negative",
			query:    PaginationQuery{Page: 0, Limit: 20},
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.query.Offset(); got != tc.expected {
				t.Errorf("expected Offset=%d, got %d", tc.expected, got)
			}
		})
	}
}

func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		limit    int
		expected int
	}{
		{name: "zero_total", total: 0, limit: 10, expected: 0},
		{name: "negative_total", total: -5, limit: 10, expected: 0},
		{name: "zero_limit", total: 50, limit: 0, expected: 0},
		{name: "negative_limit", total: 50, limit: -1, expected: 0},
		{name: "exact_pages", total: 100, limit: 10, expected: 10},
		{name: "partial_page", total: 105, limit: 10, expected: 11},
		{name: "less_than_one_page", total: 3, limit: 10, expected: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := CalculateTotalPages(tc.total, tc.limit); got != tc.expected {
				t.Errorf("expected total pages %d, got %d", tc.expected, got)
			}
		})
	}
}

func TestNewPaginationMeta(t *testing.T) {
	meta := NewPaginationMeta(2, 20, 55)
	if meta.CurrentPage != 2 {
		t.Errorf("expected CurrentPage=2, got %d", meta.CurrentPage)
	}
	if meta.PageSize != 20 {
		t.Errorf("expected PageSize=20, got %d", meta.PageSize)
	}
	if meta.TotalItems != 55 {
		t.Errorf("expected TotalItems=55, got %d", meta.TotalItems)
	}
	if meta.TotalPages != 3 {
		t.Errorf("expected TotalPages=3, got %d", meta.TotalPages)
	}
}
