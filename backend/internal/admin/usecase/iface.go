package usecase

import (
	"context"

	authEntity "github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
)

type AdminRepoIface interface {
	ListUsers(ctx context.Context, q, status string, offset, limit int) ([]authEntity.User, int64, error)
	GetUserByID(ctx context.Context, id int64) (authEntity.User, error)
	BanUser(ctx context.Context, id int64, reason *string) error
	UnbanUser(ctx context.Context, id int64) error
	RevokeAllTokensByUserID(ctx context.Context, userID int64) error
	SoftDeleteUserData(ctx context.Context, userID int64) error
	SoftDeleteUser(ctx context.Context, userID int64) error
}

type TxManagerIface interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
