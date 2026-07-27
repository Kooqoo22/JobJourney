package dto

import appDto "github.com/Kooqoo22/JobJourney/backend/internal/application/dto"

type TotalsResponse struct {
	All      int64            `json:"all"`
	ByStatus map[string]int64 `json:"by_status"`
}

type SummaryResponse struct {
	Totals             TotalsResponse               `json:"totals"`
	UpcomingEvents     []appDto.EventResponse       `json:"upcoming_events"`
	RecentApplications []appDto.ApplicationResponse `json:"recent_applications"`
}

type FunnelItem struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type TrendItem struct {
	Period string `json:"period"`
	Count  int64  `json:"count"`
}

type SourceItem struct {
	Source string `json:"source"`
	Count  int64  `json:"count"`
}

type AnalyticsResponse struct {
	Funnel        []FunnelItem `json:"funnel"`
	ResponseRate  float64      `json:"response_rate"`
	InterviewRate float64      `json:"interview_rate"`
	OfferRate     float64      `json:"offer_rate"`
	Trend         []TrendItem  `json:"trend"`
	BySource      []SourceItem `json:"by_source"`
}

type RatesResult struct {
	TotalApplied int64
	Responded    int64
	Interviewed  int64
	Offered      int64
}
