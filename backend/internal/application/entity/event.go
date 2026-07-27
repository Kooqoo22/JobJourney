package entity

import "time"

type ApplicationEvent struct {
	ID            int64      `db:"id"`
	ApplicationID int64      `db:"application_id"`
	UserID        int64      `db:"user_id"`
	Type          string     `db:"type"`
	Title         string     `db:"title"`
	EventAt       time.Time  `db:"event_at"`
	Notes         *string    `db:"notes"`
	RemindAt      *time.Time `db:"remind_at"`
	RemindedAt    *time.Time `db:"reminded_at"`
	StatusFrom    *string    `db:"status_from"`
	StatusTo      *string    `db:"status_to"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
}
