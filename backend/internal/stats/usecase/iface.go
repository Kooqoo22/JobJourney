package usecase

import (
	"context"

	appEntity "github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
	statsDto "github.com/Kooqoo22/JobJourney/backend/internal/stats/dto"
)

type StatsRepoIface interface {
	GetApplicationCounts(ctx context.Context, userID int64) (map[string]int64, error)
	GetUpcomingEvents(ctx context.Context, userID int64, limit int) ([]appEntity.ApplicationEvent, error)
	GetRecentApplications(ctx context.Context, userID int64, limit int) ([]appEntity.Application, error)
	GetFunnel(ctx context.Context, userID int64) ([]statsDto.FunnelItem, error)
	GetApplicationRates(ctx context.Context, userID int64) (statsDto.RatesResult, error)
	GetTrend(ctx context.Context, userID int64, period string) ([]statsDto.TrendItem, error)
	GetBySource(ctx context.Context, userID int64) ([]statsDto.SourceItem, error)
}
