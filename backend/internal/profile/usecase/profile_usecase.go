package usecase

import (
	"context"
	"errors"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/profile/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/profile/mapper"
	"github.com/Kooqoo22/JobJourney/backend/pkg/security"
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
		if !utils.IsValidTimezone(*req.Timezone) {
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

func (u *ProfileUsecase) ChangePassword(ctx context.Context, userID int64, req dto.ChangePasswordRequest) error {
	if fields := security.PasswordStrengthErrors("new_password", req.NewPassword); fields != nil {
		return utils.ErrUnprocessable("password does not meet the requirements", fields)
	}

	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return utils.ErrNotFound("user not found")
		}
		return utils.ErrInternal(err)
	}

	if !security.CheckPassword(user.PasswordHash, req.CurrentPassword) {
		return utils.ErrUnauthorized("current password is incorrect")
	}

	newHash, err := security.HashPassword(req.NewPassword)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if err := u.repo.UpdatePassword(ctx, userID, newHash); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}
