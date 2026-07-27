package usecase

import (
	"context"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
)

type ApplicationRepoIface interface {
	Insert(ctx context.Context, a *entity.Application) error
	GetByID(ctx context.Context, id, userID int64) (entity.Application, error)
	List(ctx context.Context, userID int64, f entity.ApplicationListFilter) ([]entity.Application, error)
	Update(ctx context.Context, a *entity.Application) error
	SoftDeleteApplication(ctx context.Context, id, userID int64) error
	SoftDeleteApplicationEvents(ctx context.Context, applicationID int64) error
	SoftDeleteApplicationDocuments(ctx context.Context, applicationID int64) error
	GetDeletedByID(ctx context.Context, id, userID int64) (entity.Application, error)
	RestoreApplication(ctx context.Context, id, userID int64) (entity.Application, error)
}

type TxManagerIface interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
