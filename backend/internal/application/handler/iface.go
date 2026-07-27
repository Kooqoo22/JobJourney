package handler

import (
	"context"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/dto"
)

type ApplicationUsecaseIface interface {
	CreateApplication(ctx context.Context, userID int64, userTZ string, req dto.CreateApplicationRequest) (dto.ApplicationResponse, error)
}
