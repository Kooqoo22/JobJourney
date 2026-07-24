package handler

import (
	"context"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/dto"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type ApplicationUsecaseIface interface {
	CreateApplication(ctx context.Context, userID int64, userTZ string, req dto.CreateApplicationRequest) (dto.ApplicationResponse, error)
	ListApplications(ctx context.Context, userID int64, userTZ string, q dto.ListApplicationsQuery) ([]dto.ApplicationResponse, utils.CursorMeta, error)
}
