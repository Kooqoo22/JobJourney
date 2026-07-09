package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/database"
)

type RefreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Insert(ctx context.Context, t *entity.RefreshToken) (err error) {
	exec := database.GetDBTx(ctx, r.db)
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES (:user_id, :token_hash, :expires_at)
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

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (entity.RefreshToken, error) {
	exec := database.GetDBTx(ctx, r.db)
	var t entity.RefreshToken
	query := `SELECT * FROM refresh_tokens WHERE token_hash = $1`
	if err := sqlx.GetContext(ctx, exec, &t, query, tokenHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.RefreshToken{}, entity.ErrRefreshTokenNotFound
		}
		return entity.RefreshToken{}, err
	}
	return t, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id int64) error {
	exec := database.GetDBTx(ctx, r.db)
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`
	_, err := exec.ExecContext(ctx, query, id)
	return err
}

func (r *RefreshTokenRepository) RevokeByHash(ctx context.Context, tokenHash string) error {
	exec := database.GetDBTx(ctx, r.db)
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL`
	_, err := exec.ExecContext(ctx, query, tokenHash)
	return err
}

func (r *RefreshTokenRepository) RevokeAllByUser(ctx context.Context, userID int64) error {
	exec := database.GetDBTx(ctx, r.db)
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`
	_, err := exec.ExecContext(ctx, query, userID)
	return err
}
