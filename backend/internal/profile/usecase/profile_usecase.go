package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/profile/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/profile/mapper"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type ProfileUsecase struct {
	repo ProfileRepoIface
}

func New(repo ProfileRepoIface) *ProfileUsecase {
	return &ProfileUsecase{repo: repo}
}

func (u *ProfileUsecase) GetProfile(ctx context.Context, userID int64) (dto.ProfileResponse, error) {
	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return dto.ProfileResponse{}, utils.ErrNotFound("user not found")
		}
		return dto.ProfileResponse{}, utils.ErrInternal(err)
	}
	return mapper.ToProfileResponse(user), nil
}

func (u *ProfileUsecase) UpdateProfile(ctx context.Context, userID int64, req dto.UpdateProfileRequest) (dto.ProfileResponse, error) {
	current, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return dto.ProfileResponse{}, utils.ErrNotFound("user not found")
		}
		return dto.ProfileResponse{}, utils.ErrInternal(err)
	}

	if req.Timezone != nil {
		if !isValidTimezone(*req.Timezone) {
			return dto.ProfileResponse{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{
				{Field: "timezone", Message: "must be a valid IANA timezone"},
			})
		}
	}

	fullName := current.FullName
	if req.FullName != nil {
		fullName = *req.FullName
	}
	avatarURL := current.AvatarURL
	if req.AvatarURL != nil {
		avatarURL = req.AvatarURL
	}
	timezone := current.Timezone
	if req.Timezone != nil {
		timezone = *req.Timezone
	}

	updated, err := u.repo.Update(ctx, userID, fullName, avatarURL, timezone)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return dto.ProfileResponse{}, utils.ErrNotFound("user not found")
		}
		return dto.ProfileResponse{}, utils.ErrInternal(err)
	}
	return mapper.ToProfileResponse(updated), nil
}

func isValidTimezone(tz string) bool {
	_, err := time.LoadLocation(tz)
	return err == nil
}
