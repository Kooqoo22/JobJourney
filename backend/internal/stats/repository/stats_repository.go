package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	appEntity "github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
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
