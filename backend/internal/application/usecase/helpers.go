package usecase

import (
	"errors"
	"time"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/mapper"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func validateCreateRequest(req dto.CreateApplicationRequest) error {
	status := req.Status
	if status == "" {
		status = "applied"
	}

	if req.AppliedDate != nil && status != "wishlist" {
		t, err := time.Parse("2006-01-02", *req.AppliedDate)
		if err == nil && t.After(time.Now().UTC()) {
			return utils.ErrUnprocessable("validation failed", []utils.FieldError{
				{Field: "applied_date", Message: "cannot be in the future"},
			})
		}
	}
	return nil
}

func wrapNotFound(err error) error {
	if errors.Is(err, entity.ErrNotFound) {
		return utils.ErrNotFound("application not found")
	}
	return utils.ErrInternal(err)
}

func applyApplicationUpdates(cur entity.Application, req dto.UpdateApplicationRequest) (entity.Application, error) {
	if req.CompanyName != nil {
		cur.CompanyName = *req.CompanyName
	}
	if req.PositionTitle != nil {
		cur.PositionTitle = *req.PositionTitle
	}
	if req.JobURL != nil {
		if *req.JobURL == "" {
			cur.JobURL = nil
		} else {
			cur.JobURL = req.JobURL
		}
	}
	if req.WorkArrangement != nil {
		if *req.WorkArrangement == "" {
			cur.WorkArrangement = nil
		} else {
			cur.WorkArrangement = req.WorkArrangement
		}
	}
	if req.EmploymentType != nil {
		if *req.EmploymentType == "" {
			cur.EmploymentType = nil
		} else {
			cur.EmploymentType = req.EmploymentType
		}
	}
	if req.Location != nil {
		if *req.Location == "" {
			cur.Location = nil
		} else {
			cur.Location = req.Location
		}
	}
	if req.Source != nil {
		if *req.Source == "" {
			cur.Source = nil
		} else {
			cur.Source = req.Source
		}
	}
	if req.Status != nil {
		cur.Status = *req.Status
	}
	if req.AppliedDate != nil {
		if *req.AppliedDate == "" {
			cur.AppliedDate = nil
		} else {
			t, err := mapper.ParseAppliedDate(req.AppliedDate)
			if err != nil {
				return entity.Application{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
					{Field: "applied_date", Message: "must be a valid date in YYYY-MM-DD format"},
				})
			}
			cur.AppliedDate = t
		}
	}
	if req.SalaryMin != nil {
		if *req.SalaryMin == "" {
			cur.SalaryMin = nil
		} else {
			d, err := mapper.ParseSalary(req.SalaryMin)
			if err != nil {
				return entity.Application{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
					{Field: "salary_min", Message: "must be a valid decimal number"},
				})
			}
			cur.SalaryMin = d
		}
	}
	if req.SalaryMax != nil {
		if *req.SalaryMax == "" {
			cur.SalaryMax = nil
		} else {
			d, err := mapper.ParseSalary(req.SalaryMax)
			if err != nil {
				return entity.Application{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
					{Field: "salary_max", Message: "must be a valid decimal number"},
				})
			}
			cur.SalaryMax = d
		}
	}
	if req.Currency != nil {
		if *req.Currency == "" {
			cur.Currency = nil
		} else {
			cur.Currency = req.Currency
		}
	}
	if req.Notes != nil {
		if *req.Notes == "" {
			cur.Notes = nil
		} else {
			cur.Notes = req.Notes
		}
	}

	if cur.SalaryMin != nil && cur.SalaryMax != nil && cur.SalaryMin.GreaterThan(*cur.SalaryMax) {
		return entity.Application{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
			{Field: "salary_min", Message: "must be less than or equal to salary_max"},
		})
	}
	if cur.AppliedDate != nil && cur.Status != "wishlist" && cur.AppliedDate.After(time.Now().UTC()) {
		return entity.Application{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
			{Field: "applied_date", Message: "cannot be in the future"},
		})
	}

	return cur, nil
}
