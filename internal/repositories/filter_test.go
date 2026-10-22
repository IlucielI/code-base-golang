package repositories

import (
	"testing"
)

func TestNewPaginationFilter(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		limit        int
		defaultLimit []int
		expectedLim  int
		expectedOff  int
	}{
		{
			name:        "first_page",
			page:        1,
			limit:       20,
			expectedLim: 20,
			expectedOff: 0,
		},
		{
			name:        "second_page",
			page:        2,
			limit:       20,
			expectedLim: 20,
			expectedOff: 20,
		},
		{
			name:        "third_page",
			page:        3,
			limit:       15,
			expectedLim: 15,
			expectedOff: 30,
		},
		{
			name:        "zero_or_negative_page_defaults_to_zero_offset",
			page:        -1,
			limit:       10,
			expectedLim: 10,
			expectedOff: 0,
		},
		{
			name:        "zero_limit_defaults_to_ten",
			page:        1,
			limit:       0,
			expectedLim: 10,
			expectedOff: 0,
		},
		{
			name:         "zero_limit_uses_custom_default",
			page:         2,
			limit:        0,
			defaultLimit: []int{25},
			expectedLim:  25,
			expectedOff:  25,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPaginationFilter(tc.page, tc.limit, tc.defaultLimit...)
			if p.Limit != tc.expectedLim {
				t.Errorf("expected limit %d, got %d", tc.expectedLim, p.Limit)
			}
			if p.Offset != tc.expectedOff {
				t.Errorf("expected offset %d, got %d", tc.expectedOff, p.Offset)
			}
		})
	}
}

func TestSortFilter_OrderClause(t *testing.T) {
	tests := []struct {
		name     string
		filter   SortFilter
		fallback []string
		expected string
	}{
		{
			name: "explicit_asc",
			filter: SortFilter{
				SortBy: "name",
				Order:  "asc",
			},
			expected: "name ASC",
		},
		{
			name: "explicit_desc",
			filter: SortFilter{
				SortBy: "created_at",
				Order:  "DESC",
			},
			expected: "created_at DESC",
		},
		{
			name: "default_asc_when_order_empty",
			filter: SortFilter{
				SortBy: "id",
				Order:  "",
			},
			expected: "id ASC",
		},
		{
			name: "fallback_when_sortBy_empty",
			filter: SortFilter{
				SortBy: "",
				Order:  "desc",
			},
			fallback: []string{"created_at DESC"},
			expected: "created_at DESC",
		},
		{
			name: "empty_when_both_empty",
			filter: SortFilter{
				SortBy: "",
				Order:  "",
			},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clause := tc.filter.OrderClause(tc.fallback...)
			if clause != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, clause)
			}
		})
	}
}
