package usecase

import (
	"context"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
)

type ApplicationRepoIface interface {
	Insert(ctx context.Context, a *entity.Application) error
	GetByID(ctx context.Context, id, userID int64) (entity.Application, error)
	GetDeletedByID(ctx context.Context, id, userID int64) (entity.Application, error)
	List(ctx context.Context, userID int64, f entity.ApplicationListFilter) ([]entity.Application, int64, error)
	Update(ctx context.Context, a *entity.Application) error
	UpdateStatus(ctx context.Context, id, userID int64, status string) (entity.Application, error)
	SetArchived(ctx context.Context, id, userID int64, isArchived bool) error
	SoftDeleteApplication(ctx context.Context, id, userID int64) error
	SoftDeleteApplicationEvents(ctx context.Context, applicationID int64) error
	RestoreApplication(ctx context.Context, id, userID int64) (entity.Application, error)
}

type EventRepoIface interface {
	InsertEvent(ctx context.Context, e *entity.ApplicationEvent) error
	GetEventByID(ctx context.Context, eventID, applicationID, userID int64) (entity.ApplicationEvent, error)
	UpdateEvent(ctx context.Context, e *entity.ApplicationEvent) error
	SoftDeleteEvent(ctx context.Context, id, applicationID, userID int64) error
	ListEvents(ctx context.Context, applicationID int64, offset, limit int) ([]entity.ApplicationEvent, int64, error)
}

type TxManagerIface interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
