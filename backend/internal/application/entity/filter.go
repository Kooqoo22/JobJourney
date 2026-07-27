package entity

type ApplicationListFilter struct {
	Keyword         string
	Status          string
	Source          string
	WorkArrangement string
	EmploymentType  string
	FromDate        string
	ToDate          string
	IsArchived      *bool
	SortBy          string
	SortDir         string
	CursorSortVal   *string
	CursorID        *int64
	Limit           int
}
