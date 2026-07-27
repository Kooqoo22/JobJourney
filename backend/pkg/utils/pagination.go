package utils

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type PageMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

func NormalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func NormalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

func NewPageMeta(total int64, page, limit int) PageMeta {
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}
	return PageMeta{Total: total, Page: page, Limit: limit, TotalPages: totalPages}
}
