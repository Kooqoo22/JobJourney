package usecase

import (
	"context"
	"time"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/mapper"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func (u *ApplicationUsecase) UpdateApplication(ctx context.Context, id, userID int64, userTZ string, req dto.UpdateApplicationRequest) (dto.ApplicationResponse, error) {
	current, err := u.repo.GetByID(ctx, id, userID)
	if err != nil {
		return dto.ApplicationResponse{}, wrapNotFound(err)
	}

	if req.UpdatedAt != nil {
		clientTime, parseErr := time.Parse(time.RFC3339, *req.UpdatedAt)
		if parseErr != nil || !current.UpdatedAt.UTC().Equal(clientTime.UTC()) {
			return dto.ApplicationResponse{}, utils.ErrConflict("application was modified by another request, please refresh and retry")
		}
	}

	updated, err := applyApplicationUpdates(current, req)
	if err != nil {
		return dto.ApplicationResponse{}, err
	}

	if err := u.repo.Update(ctx, &updated); err != nil {
		return dto.ApplicationResponse{}, utils.ErrInternal(err)
	}
	return mapper.ToApplicationResponse(updated, userTZ), nil
}

func (u *ApplicationUsecase) ListApplications(ctx context.Context, userID int64, userTZ string, q dto.ListApplicationsQuery) ([]dto.ApplicationResponse, utils.PageMeta, error) {
	page := utils.NormalizePage(q.Page)
	limit := utils.NormalizeLimit(q.Limit)
	offset := (page - 1) * limit

	f := entity.ApplicationListFilter{
		Keyword:         q.Q,
		Status:          q.Status,
		Source:          q.Source,
		WorkArrangement: q.WorkArrangement,
		EmploymentType:  q.EmploymentType,
		FromDate:        q.FromDate,
		ToDate:          q.ToDate,
		IsArchived:      q.IsArchived,
		SortBy:          q.SortBy,
		SortDir:         q.SortDir,
		Offset:          offset,
		Limit:           limit,
	}

	apps, total, err := u.repo.List(ctx, userID, f)
	if err != nil {
		return nil, utils.PageMeta{}, utils.ErrInternal(err)
	}

	responses := make([]dto.ApplicationResponse, len(apps))
	for i, a := range apps {
		responses[i] = mapper.ToApplicationResponse(a, userTZ)
	}
	return responses, utils.NewPageMeta(total, page, limit), nil
}

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

func (u *ApplicationUsecase) GetApplication(ctx context.Context, id, userID int64, userTZ string) (dto.ApplicationResponse, error) {
	a, err := u.repo.GetByID(ctx, id, userID)
	if err != nil {
		return dto.ApplicationResponse{}, wrapNotFound(err)
	}
	return mapper.ToApplicationResponse(a, userTZ), nil
}

