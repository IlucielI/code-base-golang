package repositories

// PaginationFilter encapsulates database pagination parameters.
type PaginationFilter struct {
	Limit  int
	Offset int
}

// SortFilter encapsulates database sorting parameters.
type SortFilter struct {
	SortBy string // e.g. "created_at"
	Order  string // "asc" or "desc"
}
