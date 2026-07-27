package dto

type ApplicationResponse struct {
	ID              int64   `json:"id"`
	CompanyName     string  `json:"company_name"`
	PositionTitle   string  `json:"position_title"`
	JobURL          *string `json:"job_url"`
	WorkArrangement *string `json:"work_arrangement"`
	EmploymentType  *string `json:"employment_type"`
	Location        *string `json:"location"`
	Source          *string `json:"source"`
	Status          string  `json:"status"`
	AppliedDate     *string `json:"applied_date"`
	SalaryMin       *string `json:"salary_min"`
	SalaryMax       *string `json:"salary_max"`
	Currency        *string `json:"currency"`
	Notes           *string `json:"notes"`
	IsArchived      bool    `json:"is_archived"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}
