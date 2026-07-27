package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	appEntity "github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
	statsDto "github.com/Kooqoo22/JobJourney/backend/internal/stats/dto"
)

type StatsRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

type statusCount struct {
	Status string `db:"status"`
	Count  int64  `db:"count"`
}

func (r *StatsRepository) GetApplicationCounts(ctx context.Context, userID int64) (map[string]int64, error) {
	var rows []statusCount
	query := `
		SELECT status, COUNT(*) AS count
		FROM job_applications
		WHERE user_id = $1 AND deleted_at IS NULL AND is_archived = false
		GROUP BY status`
	if err := r.db.SelectContext(ctx, &rows, query, userID); err != nil {
		return nil, err
	}
	counts := make(map[string]int64, len(rows))
	for _, rc := range rows {
		counts[rc.Status] = rc.Count
	}
	return counts, nil
}

func (r *StatsRepository) GetUpcomingEvents(ctx context.Context, userID int64, limit int) ([]appEntity.ApplicationEvent, error) {
	var events []appEntity.ApplicationEvent
	query := `
		SELECT ae.*
		FROM application_events ae
		JOIN job_applications ja ON ja.id = ae.application_id
		WHERE ae.user_id = $1
		  AND ae.deleted_at IS NULL
		  AND ja.deleted_at IS NULL
		  AND ae.event_at BETWEEN NOW() AND NOW() + INTERVAL '7 days'
		ORDER BY ae.event_at ASC, ae.id ASC
		LIMIT $2`
	if err := r.db.SelectContext(ctx, &events, query, userID, limit); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *StatsRepository) GetRecentApplications(ctx context.Context, userID int64, limit int) ([]appEntity.Application, error) {
	var apps []appEntity.Application
	query := `
		SELECT * FROM job_applications
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC, id DESC
		LIMIT $2`
	if err := r.db.SelectContext(ctx, &apps, query, userID, limit); err != nil {
		return nil, err
	}
	return apps, nil
}

type rateRow struct {
	TotalApplied int64 `db:"total_applied"`
	Responded    int64 `db:"responded"`
	Interviewed  int64 `db:"interviewed"`
	Offered      int64 `db:"offered"`
}

func (r *StatsRepository) GetApplicationRates(ctx context.Context, userID int64) (statsDto.RatesResult, error) {
	var row rateRow
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status NOT IN ('wishlist')) AS total_applied,
			COUNT(*) FILTER (WHERE status IN ('screening', 'interview', 'offer', 'accepted', 'rejected', 'withdrawn')) AS responded,
			COUNT(*) FILTER (WHERE status IN ('interview', 'offer', 'accepted')) AS interviewed,
			COUNT(*) FILTER (WHERE status IN ('offer', 'accepted')) AS offered
		FROM job_applications
		WHERE user_id = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &row, query, userID); err != nil {
		return statsDto.RatesResult{}, err
	}
	return statsDto.RatesResult{
		TotalApplied: row.TotalApplied,
		Responded:    row.Responded,
		Interviewed:  row.Interviewed,
		Offered:      row.Offered,
	}, nil
}

func (r *StatsRepository) GetFunnel(ctx context.Context, userID int64) ([]statsDto.FunnelItem, error) {
	var items []statsDto.FunnelItem
	query := `
		SELECT status, COUNT(*) AS count
		FROM job_applications
		WHERE user_id = $1 AND deleted_at IS NULL
		GROUP BY status
		ORDER BY
			CASE status
				WHEN 'wishlist' THEN 1
				WHEN 'applied' THEN 2
				WHEN 'screening' THEN 3
				WHEN 'interview' THEN 4
				WHEN 'offer' THEN 5
				WHEN 'accepted' THEN 6
				WHEN 'rejected' THEN 7
				WHEN 'withdrawn' THEN 8
				WHEN 'ghosted' THEN 9
				ELSE 10
			END`
	if err := r.db.SelectContext(ctx, &items, query, userID); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *StatsRepository) GetTrend(ctx context.Context, userID int64, period string) ([]statsDto.TrendItem, error) {
	var format string
	var interval string
	switch period {
	case "week":
		format = `TO_CHAR(created_at AT TIME ZONE 'UTC', 'IYYY-"W"IW')`
		interval = "13 weeks"
	case "quarter":
		format = `TO_CHAR(created_at AT TIME ZONE 'UTC', 'YYYY-"Q"Q')`
		interval = "4 quarters"
	case "year":
		format = `TO_CHAR(created_at AT TIME ZONE 'UTC', 'YYYY')`
		interval = "5 years"
	default:
		format = `TO_CHAR(created_at AT TIME ZONE 'UTC', 'YYYY-MM')`
		interval = "12 months"
	}
	query := fmt.Sprintf(`
		SELECT %s AS period, COUNT(*) AS count
		FROM job_applications
		WHERE user_id = $1 AND deleted_at IS NULL AND created_at >= NOW() - INTERVAL '%s'
		GROUP BY period
		ORDER BY period`, format, interval)

	var items []statsDto.TrendItem
	if err := r.db.SelectContext(ctx, &items, query, userID); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *StatsRepository) GetBySource(ctx context.Context, userID int64) ([]statsDto.SourceItem, error) {
	var items []statsDto.SourceItem
	query := `
		SELECT COALESCE(source, 'unknown') AS source, COUNT(*) AS count
		FROM job_applications
		WHERE user_id = $1 AND deleted_at IS NULL
		GROUP BY source
		ORDER BY count DESC`
	if err := r.db.SelectContext(ctx, &items, query, userID); err != nil {
		return nil, err
	}
	return items, nil
}
