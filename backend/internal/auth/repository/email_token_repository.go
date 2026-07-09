package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/database"
)

type EmailTokenRepository struct {
	db *sqlx.DB
}

func NewEmailTokenRepository(db *sqlx.DB) *EmailTokenRepository {
	return &EmailTokenRepository{db: db}
}

func (r *EmailTokenRepository) Insert(ctx context.Context, t *entity.EmailToken) (err error) {
	exec := database.GetDBTx(ctx, r.db)
	query := `
		INSERT INTO email_tokens (user_id, type, token_hash, expires_at)
		VALUES (:user_id, :type, :token_hash, :expires_at)
		RETURNING id, created_at`
	rows, err := sqlx.NamedQueryContext(ctx, exec, query, t)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := rows.Close(); err == nil {
			err = cerr
		}
	}()
	if rows.Next() {
		return rows.StructScan(t)
	}
	return rows.Err()
}

func (r *EmailTokenRepository) GetActiveByHash(ctx context.Context, tokenHash, tokenType string) (entity.EmailToken, error) {
	exec := database.GetDBTx(ctx, r.db)
	var t entity.EmailToken
	query := `SELECT * FROM email_tokens WHERE token_hash = $1 AND type = $2`
	if err := sqlx.GetContext(ctx, exec, &t, query, tokenHash, tokenType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.EmailToken{}, entity.ErrEmailTokenNotFound
		}
		return entity.EmailToken{}, err
	}
	return t, nil
}

func (r *EmailTokenRepository) MarkUsed(ctx context.Context, id int64) error {
	exec := database.GetDBTx(ctx, r.db)
	query := `UPDATE email_tokens SET used_at = NOW() WHERE id = $1 AND used_at IS NULL`
	_, err := exec.ExecContext(ctx, query, id)
	return err
}

func (r *EmailTokenRepository) InvalidateActive(ctx context.Context, userID int64, tokenType string) error {
	exec := database.GetDBTx(ctx, r.db)
	query := `UPDATE email_tokens SET used_at = NOW() WHERE user_id = $1 AND type = $2 AND used_at IS NULL`
	_, err := exec.ExecContext(ctx, query, userID, tokenType)
	return err
}
