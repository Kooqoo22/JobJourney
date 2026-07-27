package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/mapper"
	appRepo "github.com/Kooqoo22/JobJourney/backend/internal/application/repository"
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

	appliedDate, err := appRepo.ParseAppliedDate(req.AppliedDate)
	if err != nil {
		return dto.ApplicationResponse{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
			{Field: "applied_date", Message: "must be a valid date in YYYY-MM-DD format"},
		})
	}

	salaryMin, err := appRepo.ParseSalary(req.SalaryMin)
	if err != nil {
		return dto.ApplicationResponse{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
			{Field: "salary_min", Message: "must be a valid decimal number"},
		})
	}
	salaryMax, err := appRepo.ParseSalary(req.SalaryMax)
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

type appListCursor struct {
	SortVal string `json:"sv"`
	ID      int64  `json:"id"`
}

func (u *ApplicationUsecase) ListApplications(ctx context.Context, userID int64, userTZ string, q dto.ListApplicationsQuery) ([]dto.ApplicationResponse, utils.CursorMeta, error) {
	limit := utils.NormalizeLimit(q.Limit)

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
		Limit:           limit + 1,
	}

	if q.Cursor != "" {
		var c appListCursor
		if err := utils.DecodeCursor(q.Cursor, &c); err != nil {
			return nil, utils.CursorMeta{}, utils.ErrUnprocessable("invalid cursor", nil)
		}
		f.CursorSortVal = &c.SortVal
		f.CursorID = &c.ID
	}

	apps, err := u.repo.List(ctx, userID, f)
	if err != nil {
		return nil, utils.CursorMeta{}, utils.ErrInternal(err)
	}

	hasNext := len(apps) > limit
	if hasNext {
		apps = apps[:limit]
	}

	meta := utils.CursorMeta{HasNext: hasNext, Limit: limit}

	if hasNext && len(apps) > 0 {
		last := apps[len(apps)-1]
		sv := listSortVal(last, q.SortBy, q.SortDir)
		cursor, err := utils.EncodeCursor(appListCursor{SortVal: sv, ID: last.ID})
		if err != nil {
			return nil, utils.CursorMeta{}, utils.ErrInternal(err)
		}
		meta.NextCursor = cursor
	}

	responses := make([]dto.ApplicationResponse, len(apps))
	for i, a := range apps {
		responses[i] = mapper.ToApplicationResponse(a, userTZ)
	}
	return responses, meta, nil
}

func listSortVal(a entity.Application, sortBy, sortDir string) string {
	dir := strings.ToUpper(sortDir)
	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}
	switch sortBy {
	case "applied_date":
		sentinel := "1970-01-01"
		if dir == "ASC" {
			sentinel = "9999-12-31"
		}
		if a.AppliedDate == nil {
			return sentinel
		}
		return a.AppliedDate.Format("2006-01-02")
	case "company_name":
		return a.CompanyName
	case "status":
		return a.Status
	default:
		return a.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
}

func (u *ApplicationUsecase) ToggleArchive(ctx context.Context, id, userID int64, isArchived bool) error {
	if err := u.repo.SetArchived(ctx, id, userID, isArchived); err != nil {
		return wrapNotFound(err)
	}
	return nil
}

func (u *ApplicationUsecase) ChangeStatus(ctx context.Context, id, userID int64, userTZ string, req dto.ChangeStatusRequest) (dto.ApplicationResponse, error) {
	current, err := u.repo.GetByID(ctx, id, userID)
	if err != nil {
		return dto.ApplicationResponse{}, wrapNotFound(err)
	}

	if current.Status == req.Status {
		return mapper.ToApplicationResponse(current, userTZ), nil
	}

	oldStatus := current.Status
	var updated entity.Application

	err = u.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		a, txErr := u.repo.UpdateStatus(txCtx, id, userID, req.Status)
		if txErr != nil {
			return txErr
		}
		updated = a

		title := "Status changed to " + req.Status
		event := entity.ApplicationEvent{
			ApplicationID: id,
			UserID:        userID,
			Type:          "status_changed",
			Title:         title,
			EventAt:       a.UpdatedAt,
			StatusFrom:    &oldStatus,
			StatusTo:      &req.Status,
		}
		return u.eventRepo.InsertEvent(txCtx, &event)
	})
	if err != nil {
		if errors.Is(err, appRepo.ErrNotFound) {
			return dto.ApplicationResponse{}, utils.ErrNotFound("application not found")
		}
		return dto.ApplicationResponse{}, utils.ErrInternal(err)
	}

	return mapper.ToApplicationResponse(updated, userTZ), nil
}

func (u *ApplicationUsecase) RestoreApplication(ctx context.Context, id, userID int64, userTZ string) (dto.ApplicationResponse, error) {
	a, err := u.repo.GetDeletedByID(ctx, id, userID)
	if err != nil {
		return dto.ApplicationResponse{}, utils.ErrNotFound("application not found")
	}

	if a.DeletedAt != nil && time.Since(*a.DeletedAt) > 30*24*time.Hour {
		return dto.ApplicationResponse{}, utils.ErrConflict("application cannot be restored after the retention period")
	}

	restored, err := u.repo.RestoreApplication(ctx, id, userID)
	if err != nil {
		return dto.ApplicationResponse{}, utils.ErrInternal(err)
	}
	return mapper.ToApplicationResponse(restored, userTZ), nil
}

func (u *ApplicationUsecase) DeleteApplication(ctx context.Context, id, userID int64) error {
	return u.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.repo.SoftDeleteApplication(txCtx, id, userID); err != nil {
			if errors.Is(err, appRepo.ErrNotFound) {
				return utils.ErrNotFound("application not found")
			}
			return utils.ErrInternal(err)
		}
		if err := u.repo.SoftDeleteApplicationEvents(txCtx, id); err != nil {
			return utils.ErrInternal(err)
		}
		if err := u.repo.SoftDeleteApplicationDocuments(txCtx, id); err != nil {
			return utils.ErrInternal(err)
		}
		return nil
	})
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
			t, err := appRepo.ParseAppliedDate(req.AppliedDate)
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
			d, err := appRepo.ParseSalary(req.SalaryMin)
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
			d, err := appRepo.ParseSalary(req.SalaryMax)
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

func (u *ApplicationUsecase) GetApplication(ctx context.Context, id, userID int64, userTZ string) (dto.ApplicationResponse, error) {
	a, err := u.repo.GetByID(ctx, id, userID)
	if err != nil {
		return dto.ApplicationResponse{}, wrapNotFound(err)
	}
	return mapper.ToApplicationResponse(a, userTZ), nil
}

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
	if errors.Is(err, appRepo.ErrNotFound) {
		return utils.ErrNotFound("application not found")
	}
	return utils.ErrInternal(err)
}
