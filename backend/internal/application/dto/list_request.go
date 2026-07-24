package dto

type ListApplicationsQuery struct {
	Q               string `form:"q"`
	Status          string `form:"status"`
	Source          string `form:"source"`
	WorkArrangement string `form:"work_arrangement"`
	EmploymentType  string `form:"employment_type"`
	FromDate        string `form:"from_date"`
	ToDate          string `form:"to_date"`
	IsArchived      *bool  `form:"is_archived"`
	SortBy          string `form:"sort_by"`
	SortDir         string `form:"sort_dir"`
	Cursor          string `form:"cursor"`
	Limit           int    `form:"limit"`
}
