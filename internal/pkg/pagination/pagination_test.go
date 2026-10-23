package pagination_test

import (
	"testing"

	"code-base-golang/internal/pkg/pagination"
)

func TestQuery_SetDefaults(t *testing.T) {
	t.Run("empty query gets standard defaults", func(t *testing.T) {
		q := pagination.Query{}
		q.SetDefaults()
		if q.Page != 1 {
			t.Errorf("expected page 1, got %d", q.Page)
		}
		if q.Limit != pagination.DefaultLimit {
			t.Errorf("expected limit %d, got %d", pagination.DefaultLimit, q.Limit)
		}
	})

	t.Run("custom default limit", func(t *testing.T) {
		q := pagination.Query{}
		q.SetDefaults(50)
		if q.Limit != 50 {
			t.Errorf("expected limit 50, got %d", q.Limit)
		}
	})

	t.Run("limit exceeds maximum", func(t *testing.T) {
		q := pagination.Query{Page: 2, Limit: 200}
		q.SetDefaults()
		if q.Limit != pagination.MaxLimit {
			t.Errorf("expected limit %d, got %d", pagination.MaxLimit, q.Limit)
		}
	})
}

func TestQuery_Offset(t *testing.T) {
	cases := []struct {
		page   int
		limit  int
		offset int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 20, 40},
		{0, 10, 0},
	}

	for _, tc := range cases {
		q := pagination.Query{Page: tc.page, Limit: tc.limit}
		if got := q.Offset(); got != tc.offset {
			t.Errorf("for page %d limit %d expected offset %d, got %d", tc.page, tc.limit, tc.offset, got)
		}
	}
}

func TestCalculateTotalPages(t *testing.T) {
	cases := []struct {
		total int64
		limit int
		pages int
	}{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{25, 10, 3},
		{100, 0, 0},
	}

	for _, tc := range cases {
		if got := pagination.CalculateTotalPages(tc.total, tc.limit); got != tc.pages {
			t.Errorf("for total %d limit %d expected %d pages, got %d", tc.total, tc.limit, tc.pages, got)
		}
	}
}

func TestNewMeta(t *testing.T) {
	meta := pagination.NewMeta(2, 15, 40)
	if meta.CurrentPage != 2 {
		t.Errorf("expected current page 2, got %d", meta.CurrentPage)
	}
	if meta.PageSize != 15 {
		t.Errorf("expected page size 15, got %d", meta.PageSize)
	}
	if meta.TotalItems != 40 {
		t.Errorf("expected total items 40, got %d", meta.TotalItems)
	}
	if meta.TotalPages != 3 {
		t.Errorf("expected total pages 3, got %d", meta.TotalPages)
	}
}
