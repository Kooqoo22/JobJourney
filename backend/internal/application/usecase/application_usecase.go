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
