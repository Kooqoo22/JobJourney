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
