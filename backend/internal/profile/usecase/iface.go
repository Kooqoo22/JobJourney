package usecase

import (
	"context"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
)

type ProfileRepoIface interface {
	GetByID(ctx context.Context, userID int64) (entity.User, error)
	Update(ctx context.Context, userID int64, fullName string, avatarURL *string, timezone string) (entity.User, error)
}
