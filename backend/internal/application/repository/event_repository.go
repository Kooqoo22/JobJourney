package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/database"
)

type EventRepository struct {
	db *sqlx.DB
}

func NewEventRepository(db *sqlx.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) ListEvents(ctx context.Context, applicationID int64, offset, limit int) ([]entity.ApplicationEvent, int64, error) {
	exec := database.GetDBTx(ctx, r.db)

	var total int64
	countQuery := `SELECT COUNT(*) FROM application_events WHERE application_id = $1 AND deleted_at IS NULL`
	if err := sqlx.GetContext(ctx, exec, &total, countQuery, applicationID); err != nil {
		return nil, 0, err
	}

	var events []entity.ApplicationEvent
	dataQuery := `
		SELECT * FROM application_events
		WHERE application_id = $1 AND deleted_at IS NULL
		ORDER BY event_at ASC, id ASC
		LIMIT $2 OFFSET $3`
	if err := sqlx.SelectContext(ctx, exec, &events, dataQuery, applicationID, limit, offset); err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *EventRepository) InsertEvent(ctx context.Context, e *entity.ApplicationEvent) (err error) {
	exec := database.GetDBTx(ctx, r.db)
	query := `
		INSERT INTO application_events
			(application_id, user_id, type, title, event_at, notes, remind_at, status_from, status_to)
		VALUES
			(:application_id, :user_id, :type, :title, :event_at, :notes, :remind_at, :status_from, :status_to)
		RETURNING id, created_at, updated_at`
	rows, err := sqlx.NamedQueryContext(ctx, exec, query, e)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := rows.Close(); err == nil {
			err = cerr
		}
	}()
	if rows.Next() {
		return rows.StructScan(e)
	}
	return rows.Err()
}
