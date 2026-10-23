package pagination

// Default and maximum limits for pagination queries.
const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Query represents standard HTTP query parameters for pagination.
type Query struct {
	Page  int `form:"page" json:"page"`
	Limit int `form:"limit" json:"limit"`
}

// SetDefaults applies default values if Page or Limit are missing or invalid.
// An optional custom default limit can be supplied.
func (p *Query) SetDefaults(defaultLimit ...int) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		if len(defaultLimit) > 0 && defaultLimit[0] > 0 {
			p.Limit = defaultLimit[0]
		} else {
			p.Limit = DefaultLimit
		}
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
}

// Offset calculates the database offset based on Page and Limit.
func (p *Query) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit
}

// Meta encapsulates pagination metadata returned to clients.
type Meta struct {
	CurrentPage int   `json:"current_page"`
	PageSize    int   `json:"page_size"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
}

// CalculateTotalPages calculates total pages for a given total item count and limit.
func CalculateTotalPages(total int64, limit int) int {
	if limit <= 0 || total <= 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}

// NewMeta creates a populated Meta instance.
func NewMeta(page, limit int, total int64) Meta {
	return Meta{
		CurrentPage: page,
		PageSize:    limit,
		TotalItems:  total,
		TotalPages:  CalculateTotalPages(total, limit),
	}
}

// Filter encapsulates database pagination parameters (Limit and Offset).
type Filter struct {
	Limit  int
	Offset int
}

// Sort encapsulates database sorting parameters.
type Sort struct {
	SortBy string // e.g. "created_at"
	Order  string // "asc" or "desc"
}
