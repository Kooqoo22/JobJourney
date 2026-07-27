package dto

type UpdateApplicationRequest struct {
	CompanyName     *string `json:"company_name" binding:"omitempty,max=200"`
	PositionTitle   *string `json:"position_title" binding:"omitempty,max=200"`
	JobURL          *string `json:"job_url" binding:"omitempty"`
	WorkArrangement *string `json:"work_arrangement" binding:"omitempty,oneof=remote onsite hybrid"`
	EmploymentType  *string `json:"employment_type" binding:"omitempty,oneof=full_time part_time contract internship freelance"`
	Location        *string `json:"location" binding:"omitempty"`
	Source          *string `json:"source" binding:"omitempty,max=200"`
	Status          *string `json:"status" binding:"omitempty,oneof=wishlist applied screening interview offer accepted rejected withdrawn ghosted"`
	AppliedDate     *string `json:"applied_date" binding:"omitempty"`
	SalaryMin       *string `json:"salary_min" binding:"omitempty"`
	SalaryMax       *string `json:"salary_max" binding:"omitempty"`
	Currency        *string `json:"currency" binding:"omitempty,max=10"`
	Notes           *string `json:"notes" binding:"omitempty,max=5000"`
	UpdatedAt       *string `json:"updated_at" binding:"omitempty"`
}
