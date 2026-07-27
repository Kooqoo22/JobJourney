package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/database"
)

type ProfileRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) GetByID(ctx context.Context, userID int64) (entity.User, error) {
	exec := database.GetDBTx(ctx, r.db)
	var u entity.User
	query := `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`
	if err := sqlx.GetContext(ctx, exec, &u, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}
		return entity.User{}, err
	}
	return u, nil
}

func (r *ProfileRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	exec := database.GetDBTx(ctx, r.db)
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	_, err := exec.ExecContext(ctx, query, passwordHash, userID)
	return err
}

func (r *ProfileRepository) Update(ctx context.Context, userID int64, fullName string, avatarURL *string, timezone string) (entity.User, error) {
	exec := database.GetDBTx(ctx, r.db)
	var u entity.User
	query := `
		UPDATE users SET full_name = $1, avatar_url = $2, timezone = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
		RETURNING *`
	if err := sqlx.GetContext(ctx, exec, &u, query, fullName, avatarURL, timezone, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}
		return entity.User{}, err
	}
	return u, nil
}
