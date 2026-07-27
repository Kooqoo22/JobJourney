package mapper

import (
	"time"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func ToApplicationResponse(a entity.Application, tz string) dto.ApplicationResponse {
	resp := dto.ApplicationResponse{
		ID:              a.ID,
		CompanyName:     a.CompanyName,
		PositionTitle:   a.PositionTitle,
		JobURL:          a.JobURL,
		WorkArrangement: a.WorkArrangement,
		EmploymentType:  a.EmploymentType,
		Location:        a.Location,
		Source:          a.Source,
		Status:          a.Status,
		Currency:        a.Currency,
		Notes:           a.Notes,
		IsArchived:      a.IsArchived,
		CreatedAt:       utils.ToLocal(a.CreatedAt, tz).Format(time.RFC3339),
		UpdatedAt:       utils.ToLocal(a.UpdatedAt, tz).Format(time.RFC3339),
	}
	if a.AppliedDate != nil {
		d := a.AppliedDate.Format("2006-01-02")
		resp.AppliedDate = &d
	}
	if a.SalaryMin != nil {
		s := a.SalaryMin.String()
		resp.SalaryMin = &s
	}
	if a.SalaryMax != nil {
		s := a.SalaryMax.String()
		resp.SalaryMax = &s
	}
	return resp
}
