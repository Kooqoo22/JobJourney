package dto

import appDto "github.com/Kooqoo22/JobJourney/backend/internal/application/dto"

type TotalsResponse struct {
	All      int64            `json:"all"`
	ByStatus map[string]int64 `json:"by_status"`
}

type SummaryResponse struct {
	Totals             TotalsResponse              `json:"totals"`
	UpcomingEvents     []appDto.EventResponse      `json:"upcoming_events"`
	RecentApplications []appDto.ApplicationResponse `json:"recent_applications"`
}
