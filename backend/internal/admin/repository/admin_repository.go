package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	authEntity "github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
)

type AdminRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) ListUsers(ctx context.Context, q, status string, offset, limit int) ([]authEntity.User, int64, error) {
	args := []interface{}{}
	n := 1
	where := []string{"deleted_at IS NULL"}

	if q != "" {
		like := "%" + q + "%"
		where = append(where, fmt.Sprintf("(email ILIKE $%d OR full_name ILIKE $%d)", n, n+1))
		args = append(args, like, like)
		n += 2
	}

	switch status {
	case "active":
		where = append(where, "is_banned = false")
	case "banned":
		where = append(where, "is_banned = true")
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	dataQuery := fmt.Sprintf(`
		SELECT * FROM users
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, n, n+1)

	var users []authEntity.User
	if err := r.db.SelectContext(ctx, &users, dataQuery, args...); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *AdminRepository) GetUserByID(ctx context.Context, id int64) (authEntity.User, error) {
	var u authEntity.User
	if err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return authEntity.User{}, authEntity.ErrUserNotFound
		}
		return authEntity.User{}, err
	}
	return u, nil
}
