package handler

import (
	"context"

	statsDto "github.com/Kooqoo22/JobJourney/backend/internal/stats/dto"
)

type StatsUsecaseIface interface {
	GetSummary(ctx context.Context, userID int64, userTZ string) (statsDto.SummaryResponse, error)
	GetAnalytics(ctx context.Context, userID int64, period string) (statsDto.AnalyticsResponse, error)
}
