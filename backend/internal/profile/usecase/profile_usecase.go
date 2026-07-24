package usecase

import (
	"context"
	"errors"
	"time"

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

func (u *ProfileUsecase) ChangePassword(ctx context.Context, userID int64, req dto.ChangePasswordRequest) error {
	if fields := passwordStrengthErrors(req.NewPassword); fields != nil {
		return utils.ErrUnprocessable("password does not meet the requirements", fields)
	}

	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return utils.ErrNotFound("user not found")
		}
		return utils.ErrInternal(err)
	}

	if user.AuthProvider != "local" || user.PasswordHash == nil {
		return utils.ErrForbidden("password cannot be changed for accounts that use Google sign-in")
	}
	if !security.CheckPassword(*user.PasswordHash, req.CurrentPassword) {
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

func passwordStrengthErrors(pw string) []utils.FieldError {
	var hasLetter, hasDigit bool
	for _, r := range pw {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
			hasLetter = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	var out []utils.FieldError
	if len([]rune(pw)) < 8 {
		out = append(out, utils.FieldError{Field: "new_password", Message: "must be at least 8 characters"})
	}
	if !hasLetter || !hasDigit {
		out = append(out, utils.FieldError{Field: "new_password", Message: "must contain both letters and numbers"})
	}
	return out
}

func isValidTimezone(tz string) bool {
	_, err := time.LoadLocation(tz)
	return err == nil
}
