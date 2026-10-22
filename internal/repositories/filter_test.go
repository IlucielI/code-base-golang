package repositories

import (
	"testing"
)

func TestFilterStructs(t *testing.T) {
	t.Run("PaginationFilter", func(t *testing.T) {
		p := PaginationFilter{
			Limit:  20,
			Offset: 40,
		}
		if p.Limit != 20 || p.Offset != 40 {
			t.Fatalf("unexpected fields in PaginationFilter: %+v", p)
		}
	})

	t.Run("SortFilter", func(t *testing.T) {
		s := SortFilter{
			SortBy: "created_at",
			Order:  "desc",
		}
		if s.SortBy != "created_at" || s.Order != "desc" {
			t.Fatalf("unexpected fields in SortFilter: %+v", s)
		}
	})
}
