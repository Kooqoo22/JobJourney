package handler

import (
	"context"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/dto"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type ApplicationUsecaseIface interface {
	CreateApplication(ctx context.Context, userID int64, userTZ string, req dto.CreateApplicationRequest) (dto.ApplicationResponse, error)
	ListApplications(ctx context.Context, userID int64, userTZ string, q dto.ListApplicationsQuery) ([]dto.ApplicationResponse, utils.PageMeta, error)
	GetApplication(ctx context.Context, id, userID int64, userTZ string) (dto.ApplicationResponse, error)
	UpdateApplication(ctx context.Context, id, userID int64, userTZ string, req dto.UpdateApplicationRequest) (dto.ApplicationResponse, error)
	DeleteApplication(ctx context.Context, id, userID int64) error
	RestoreApplication(ctx context.Context, id, userID int64, userTZ string) (dto.ApplicationResponse, error)
	ChangeStatus(ctx context.Context, id, userID int64, userTZ string, req dto.ChangeStatusRequest) (dto.ApplicationResponse, error)
	ToggleArchive(ctx context.Context, id, userID int64, isArchived bool) error
	CreateEvent(ctx context.Context, applicationID, userID int64, userTZ string, req dto.CreateEventRequest) (dto.EventResponse, error)
}
