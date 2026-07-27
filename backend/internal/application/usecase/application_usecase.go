package usecase

import (
	"context"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/mapper"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type ApplicationUsecase struct {
	repo ApplicationRepoIface
	tx   TxManagerIface
}

func New(repo ApplicationRepoIface, tx TxManagerIface) *ApplicationUsecase {
	return &ApplicationUsecase{repo: repo, tx: tx}
}

func (u *ApplicationUsecase) CreateApplication(ctx context.Context, userID int64, userTZ string, req dto.CreateApplicationRequest) (dto.ApplicationResponse, error) {
	if err := validateCreateRequest(req); err != nil {
		return dto.ApplicationResponse{}, err
	}

	status := req.Status
	if status == "" {
		status = "applied"
	}

	appliedDate, err := mapper.ParseAppliedDate(req.AppliedDate)
	if err != nil {
		return dto.ApplicationResponse{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
			{Field: "applied_date", Message: "must be a valid date in YYYY-MM-DD format"},
		})
	}

	salaryMin, err := mapper.ParseSalary(req.SalaryMin)
	if err != nil {
		return dto.ApplicationResponse{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
			{Field: "salary_min", Message: "must be a valid decimal number"},
		})
	}
	salaryMax, err := mapper.ParseSalary(req.SalaryMax)
	if err != nil {
		return dto.ApplicationResponse{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
			{Field: "salary_max", Message: "must be a valid decimal number"},
		})
	}

	if salaryMin != nil && salaryMax != nil && salaryMin.GreaterThan(*salaryMax) {
		return dto.ApplicationResponse{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
			{Field: "salary_min", Message: "must be less than or equal to salary_max"},
		})
	}

	a := entity.Application{
		UserID:          userID,
		CompanyName:     req.CompanyName,
		PositionTitle:   req.PositionTitle,
		JobURL:          req.JobURL,
		WorkArrangement: req.WorkArrangement,
		EmploymentType:  req.EmploymentType,
		Location:        req.Location,
		Source:          req.Source,
		Status:          status,
		AppliedDate:     appliedDate,
		SalaryMin:       salaryMin,
		SalaryMax:       salaryMax,
		Currency:        req.Currency,
		Notes:           req.Notes,
		IsArchived:      false,
	}

	if err := u.repo.Insert(ctx, &a); err != nil {
		return dto.ApplicationResponse{}, utils.ErrInternal(err)
	}

	return mapper.ToApplicationResponse(a, userTZ), nil
}

