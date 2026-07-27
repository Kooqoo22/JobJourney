package usecase

import (
	"context"

	appEntity "github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
)

type StatsRepoIface interface {
	GetApplicationCounts(ctx context.Context, userID int64) (map[string]int64, error)
	GetUpcomingEvents(ctx context.Context, userID int64, limit int) ([]appEntity.ApplicationEvent, error)
	GetRecentApplications(ctx context.Context, userID int64, limit int) ([]appEntity.Application, error)
}
