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
