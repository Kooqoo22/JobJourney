package usecase

import (
	"context"

	appMapper "github.com/Kooqoo22/JobJourney/backend/internal/application/mapper"
	statsDto "github.com/Kooqoo22/JobJourney/backend/internal/stats/dto"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type StatsUsecase struct {
	repo StatsRepoIface
}

func New(repo StatsRepoIface) *StatsUsecase {
	return &StatsUsecase{repo: repo}
}

func (u *StatsUsecase) GetSummary(ctx context.Context, userID int64, userTZ string) (statsDto.SummaryResponse, error) {
	counts, err := u.repo.GetApplicationCounts(ctx, userID)
	if err != nil {
		return statsDto.SummaryResponse{}, utils.ErrInternal(err)
	}

	var total int64
	for _, c := range counts {
		total += c
	}

	upcomingEvents, err := u.repo.GetUpcomingEvents(ctx, userID, 10)
	if err != nil {
		return statsDto.SummaryResponse{}, utils.ErrInternal(err)
	}

	recentApps, err := u.repo.GetRecentApplications(ctx, userID, 5)
	if err != nil {
		return statsDto.SummaryResponse{}, utils.ErrInternal(err)
	}

	summary := statsDto.SummaryResponse{
		Totals: statsDto.TotalsResponse{
			All:      total,
			ByStatus: counts,
		},
	}

	for _, e := range upcomingEvents {
		summary.UpcomingEvents = append(summary.UpcomingEvents, appMapper.ToEventResponse(e, userTZ))
	}
	for _, a := range recentApps {
		summary.RecentApplications = append(summary.RecentApplications, appMapper.ToApplicationResponse(a, userTZ))
	}

	return summary, nil
}

func (u *StatsUsecase) GetAnalytics(ctx context.Context, userID int64, period string) (statsDto.AnalyticsResponse, error) {
	validPeriods := map[string]bool{"week": true, "month": true, "quarter": true, "year": true}
	if !validPeriods[period] {
		period = "month"
	}

	funnel, err := u.repo.GetFunnel(ctx, userID)
	if err != nil {
		return statsDto.AnalyticsResponse{}, utils.ErrInternal(err)
	}

	rates, err := u.repo.GetApplicationRates(ctx, userID)
	if err != nil {
		return statsDto.AnalyticsResponse{}, utils.ErrInternal(err)
	}

	trend, err := u.repo.GetTrend(ctx, userID, period)
	if err != nil {
		return statsDto.AnalyticsResponse{}, utils.ErrInternal(err)
	}

	bySource, err := u.repo.GetBySource(ctx, userID)
	if err != nil {
		return statsDto.AnalyticsResponse{}, utils.ErrInternal(err)
	}

	var responseRate, interviewRate, offerRate float64
	if rates.TotalApplied > 0 {
		responseRate = float64(rates.Responded) / float64(rates.TotalApplied)
		interviewRate = float64(rates.Interviewed) / float64(rates.TotalApplied)
		offerRate = float64(rates.Offered) / float64(rates.TotalApplied)
	}

	if funnel == nil {
		funnel = []statsDto.FunnelItem{}
	}
	if trend == nil {
		trend = []statsDto.TrendItem{}
	}
	if bySource == nil {
		bySource = []statsDto.SourceItem{}
	}

	return statsDto.AnalyticsResponse{
		Funnel:        funnel,
		ResponseRate:  responseRate,
		InterviewRate: interviewRate,
		OfferRate:     offerRate,
		Trend:         trend,
		BySource:      bySource,
	}, nil
}
