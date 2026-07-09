package usecase

import (
	"context"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
)

type UserRepoIface interface {
	Insert(ctx context.Context, u *entity.User) error
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	GetByID(ctx context.Context, id int64) (entity.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	SetVerified(ctx context.Context, userID int64) error
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID int64) error
}

type EmailTokenRepoIface interface {
	Insert(ctx context.Context, t *entity.EmailToken) error
	GetActiveByHash(ctx context.Context, tokenHash, tokenType string) (entity.EmailToken, error)
	MarkUsed(ctx context.Context, id int64) error
	InvalidateActive(ctx context.Context, userID int64, tokenType string) error
}

type RefreshTokenRepoIface interface {
	Insert(ctx context.Context, t *entity.RefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (entity.RefreshToken, error)
	Revoke(ctx context.Context, id int64) error
	RevokeByHash(ctx context.Context, tokenHash string) error
	RevokeAllByUser(ctx context.Context, userID int64) error
}
