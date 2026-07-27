package usecase

import (
	"context"
	"time"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/mapper"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type ApplicationUsecase struct {
	repo      ApplicationRepoIface
	eventRepo EventRepoIface
	tx        TxManagerIface
}

func New(repo ApplicationRepoIface, eventRepo EventRepoIface, tx TxManagerIface) *ApplicationUsecase {
	return &ApplicationUsecase{repo: repo, eventRepo: eventRepo, tx: tx}
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

func (u *ApplicationUsecase) DeleteApplication(ctx context.Context, id, userID int64) error {
	return u.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.repo.SoftDeleteApplication(txCtx, id, userID); err != nil {
			return wrapNotFound(err)
		}
		if err := u.repo.SoftDeleteApplicationEvents(txCtx, id); err != nil {
			return utils.ErrInternal(err)
		}
		return nil
	})
}

func (u *ApplicationUsecase) RestoreApplication(ctx context.Context, id, userID int64, userTZ string) (dto.ApplicationResponse, error) {
	if _, err := u.repo.GetDeletedByID(ctx, id, userID); err != nil {
		return dto.ApplicationResponse{}, utils.ErrNotFound("application not found")
	}
	a, err := u.repo.RestoreApplication(ctx, id, userID)
	if err != nil {
		return dto.ApplicationResponse{}, utils.ErrInternal(err)
	}
	return mapper.ToApplicationResponse(a, userTZ), nil
}

func (u *ApplicationUsecase) ChangeStatus(ctx context.Context, id, userID int64, userTZ string, req dto.ChangeStatusRequest) (dto.ApplicationResponse, error) {
	a, err := u.repo.UpdateStatus(ctx, id, userID, req.Status)
	if err != nil {
		return dto.ApplicationResponse{}, wrapNotFound(err)
	}
	return mapper.ToApplicationResponse(a, userTZ), nil
}

func (u *ApplicationUsecase) ToggleArchive(ctx context.Context, id, userID int64, isArchived bool) error {
	if err := u.repo.SetArchived(ctx, id, userID, isArchived); err != nil {
		return wrapNotFound(err)
	}
	return nil
}

func (u *ApplicationUsecase) CreateEvent(ctx context.Context, applicationID, userID int64, userTZ string, req dto.CreateEventRequest) (dto.EventResponse, error) {
	if _, err := u.repo.GetByID(ctx, applicationID, userID); err != nil {
		return dto.EventResponse{}, utils.ErrNotFound("application not found")
	}

	eventAt, err := time.Parse(time.RFC3339, req.EventAt)
	if err != nil {
		return dto.EventResponse{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
			{Field: "event_at", Message: "must be a valid RFC 3339 datetime"},
		})
	}

	var remindAt *time.Time
	if req.RemindAt != nil {
		t, parseErr := time.Parse(time.RFC3339, *req.RemindAt)
		if parseErr != nil {
			return dto.EventResponse{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
				{Field: "remind_at", Message: "must be a valid RFC 3339 datetime"},
			})
		}
		if !t.After(time.Now().UTC()) {
			return dto.EventResponse{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
				{Field: "remind_at", Message: "must be in the future"},
			})
		}
		remindAt = &t
	}

	e := entity.ApplicationEvent{
		ApplicationID: applicationID,
		UserID:        userID,
		Type:          req.Type,
		Title:         req.Title,
		EventAt:       eventAt.UTC(),
		Notes:         req.Notes,
		RemindAt:      remindAt,
	}

	if err := u.eventRepo.InsertEvent(ctx, &e); err != nil {
		return dto.EventResponse{}, utils.ErrInternal(err)
	}

	return mapper.ToEventResponse(e, userTZ), nil
}

func (u *ApplicationUsecase) ListEvents(ctx context.Context, applicationID, userID int64, userTZ string, q dto.EventListQuery) ([]dto.EventResponse, utils.PageMeta, error) {
	if _, err := u.repo.GetByID(ctx, applicationID, userID); err != nil {
		return nil, utils.PageMeta{}, utils.ErrNotFound("application not found")
	}

	page := utils.NormalizePage(q.Page)
	limit := utils.NormalizeLimit(q.Limit)
	offset := (page - 1) * limit

	events, total, err := u.eventRepo.ListEvents(ctx, applicationID, offset, limit)
	if err != nil {
		return nil, utils.PageMeta{}, utils.ErrInternal(err)
	}

	responses := make([]dto.EventResponse, len(events))
	for i, e := range events {
		responses[i] = mapper.ToEventResponse(e, userTZ)
	}
	return responses, utils.NewPageMeta(total, page, limit), nil
}
