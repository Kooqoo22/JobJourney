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

func (r *EventRepository) GetEventByID(ctx context.Context, eventID, applicationID, userID int64) (entity.ApplicationEvent, error) {
	exec := database.GetDBTx(ctx, r.db)
	var e entity.ApplicationEvent
	query := `SELECT * FROM application_events WHERE id = $1 AND application_id = $2 AND user_id = $3 AND deleted_at IS NULL`
	if err := sqlx.GetContext(ctx, exec, &e, query, eventID, applicationID, userID); err != nil {
		return entity.ApplicationEvent{}, err
	}
	return e, nil
}

func (r *EventRepository) UpdateEvent(ctx context.Context, e *entity.ApplicationEvent) (err error) {
	exec := database.GetDBTx(ctx, r.db)
	query := `
		UPDATE application_events
		SET type = :type, title = :title, event_at = :event_at, notes = :notes, remind_at = :remind_at, updated_at = NOW()
		WHERE id = :id AND deleted_at IS NULL
		RETURNING updated_at`
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
		return rows.Scan(&e.UpdatedAt)
	}
	return rows.Err()
}

func (r *EventRepository) SoftDeleteEvent(ctx context.Context, id, applicationID, userID int64) error {
	exec := database.GetDBTx(ctx, r.db)
	res, err := exec.ExecContext(ctx,
		`UPDATE application_events SET deleted_at = NOW() WHERE id = $1 AND application_id = $2 AND user_id = $3 AND deleted_at IS NULL`,
		id, applicationID, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return entity.ErrEventNotFound
	}
	return nil
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
