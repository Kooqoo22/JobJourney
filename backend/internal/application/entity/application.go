package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type Application struct {
	ID              int64            `db:"id"`
	UserID          int64            `db:"user_id"`
	CompanyName     string           `db:"company_name"`
	PositionTitle   string           `db:"position_title"`
	JobURL          *string          `db:"job_url"`
	WorkArrangement *string          `db:"work_arrangement"`
	EmploymentType  *string          `db:"employment_type"`
	Location        *string          `db:"location"`
	Source          *string          `db:"source"`
	Status          string           `db:"status"`
	AppliedDate     *time.Time       `db:"applied_date"`
	SalaryMin       *decimal.Decimal `db:"salary_min"`
	SalaryMax       *decimal.Decimal `db:"salary_max"`
	Currency        *string          `db:"currency"`
	Notes           *string          `db:"notes"`
	IsArchived      bool             `db:"is_archived"`
	CreatedAt       time.Time        `db:"created_at"`
	UpdatedAt       time.Time        `db:"updated_at"`
	DeletedAt       *time.Time       `db:"deleted_at"`
}
