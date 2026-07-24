package handler

import (
	"context"

	"github.com/Kooqoo22/JobJourney/backend/internal/profile/dto"
)

type ProfileUsecaseIface interface {
	GetProfile(ctx context.Context, userID int64) (dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID int64, req dto.UpdateProfileRequest) (dto.ProfileResponse, error)
	ChangePassword(ctx context.Context, userID int64, req dto.ChangePasswordRequest) error
	UpdatePreferences(ctx context.Context, userID int64, req dto.UpdatePreferencesRequest) (dto.PreferencesResponse, error)
}
