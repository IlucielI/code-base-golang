package dtos

// Default and maximum limits for pagination queries.
const (
	DefaultPaginationLimit = 20
	MaxPaginationLimit     = 100
)

// PaginationQuery represents standard query parameters for pagination.
type PaginationQuery struct {
	Page  int `form:"page" json:"page"`
	Limit int `form:"limit" json:"limit"`
}

// SetDefaults applies default values if Page or Limit are missing or invalid.
// An optional custom default limit can be supplied.
func (p *PaginationQuery) SetDefaults(defaultLimit ...int) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		if len(defaultLimit) > 0 && defaultLimit[0] > 0 {
			p.Limit = defaultLimit[0]
		} else {
			p.Limit = DefaultPaginationLimit
		}
	}
	if p.Limit > MaxPaginationLimit {
		p.Limit = MaxPaginationLimit
	}
}

// Offset calculates the database offset based on Page and Limit.
func (p *PaginationQuery) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit
}

// PaginationMeta encapsulates pagination metadata returned to clients.
type PaginationMeta struct {
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

// NewPaginationMeta creates a populated PaginationMeta instance.
func NewPaginationMeta(page, limit int, total int64) PaginationMeta {
	return PaginationMeta{
		CurrentPage: page,
		PageSize:    limit,
		TotalItems:  total,
		TotalPages:  CalculateTotalPages(total, limit),
	}
}
