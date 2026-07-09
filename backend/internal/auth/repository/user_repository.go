package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/database"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Insert(ctx context.Context, u *entity.User) (err error) {
	exec := database.GetDBTx(ctx, r.db)
	query := `
		INSERT INTO users (email, password_hash, auth_provider, full_name, timezone, is_verified, role)
		VALUES (:email, :password_hash, :auth_provider, :full_name, :timezone, :is_verified, :role)
		RETURNING id, created_at, updated_at`
	rows, err := sqlx.NamedQueryContext(ctx, exec, query, u)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := rows.Close(); err == nil {
			err = cerr
		}
	}()
	if rows.Next() {
		return rows.StructScan(u)
	}
	return rows.Err()
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	exec := database.GetDBTx(ctx, r.db)
	var u entity.User
	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
	if err := sqlx.GetContext(ctx, exec, &u, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, ErrUserNotFound
		}
		return entity.User{}, err
	}
	return u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (entity.User, error) {
	exec := database.GetDBTx(ctx, r.db)
	var u entity.User
	query := `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`
	if err := sqlx.GetContext(ctx, exec, &u, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, ErrUserNotFound
		}
		return entity.User{}, err
	}
	return u, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exec := database.GetDBTx(ctx, r.db)
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
	if err := sqlx.GetContext(ctx, exec, &exists, query, email); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *UserRepository) SetVerified(ctx context.Context, userID int64) error {
	exec := database.GetDBTx(ctx, r.db)
	query := `UPDATE users SET is_verified = TRUE, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := exec.ExecContext(ctx, query, userID)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	exec := database.GetDBTx(ctx, r.db)
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	_, err := exec.ExecContext(ctx, query, passwordHash, userID)
	return err
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	exec := database.GetDBTx(ctx, r.db)
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := exec.ExecContext(ctx, query, userID)
	return err
}
