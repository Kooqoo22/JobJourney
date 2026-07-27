package usecase

import (
	"context"

	authEntity "github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
)

type AdminRepoIface interface {
	ListUsers(ctx context.Context, q, status string, offset, limit int) ([]authEntity.User, int64, error)
	GetUserByID(ctx context.Context, id int64) (authEntity.User, error)
}
