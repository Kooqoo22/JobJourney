package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/database"
)

type ApplicationRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) Insert(ctx context.Context, a *entity.Application) (err error) {
	exec := database.GetDBTx(ctx, r.db)
	query := `
		INSERT INTO job_applications
			(user_id, company_name, position_title, job_url, work_arrangement, employment_type,
			 location, source, status, applied_date, salary_min, salary_max, currency, notes, is_archived)
		VALUES
			(:user_id, :company_name, :position_title, :job_url, :work_arrangement, :employment_type,
			 :location, :source, :status, :applied_date, :salary_min, :salary_max, :currency, :notes, :is_archived)
		RETURNING id, created_at, updated_at`
	rows, err := sqlx.NamedQueryContext(ctx, exec, query, a)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := rows.Close(); err == nil {
			err = cerr
		}
	}()
	if rows.Next() {
		return rows.StructScan(a)
	}
	return rows.Err()
}

func (r *ApplicationRepository) GetByID(ctx context.Context, id, userID int64) (entity.Application, error) {
	exec := database.GetDBTx(ctx, r.db)
	var a entity.Application
	query := `SELECT * FROM job_applications WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	if err := sqlx.GetContext(ctx, exec, &a, query, id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Application{}, ErrNotFound
		}
		return entity.Application{}, err
	}
	return a, nil
}

var ErrNotFound = errors.New("application not found")

func parseSalary(s *string) (*decimal.Decimal, error) {
	if s == nil {
		return nil, nil
	}
	d, err := decimal.NewFromString(*s)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func ParseAppliedDate(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func ParseSalary(s *string) (*decimal.Decimal, error) {
	return parseSalary(s)
}
