package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

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
			return entity.Application{}, entity.ErrNotFound
		}
		return entity.Application{}, err
	}
	return a, nil
}

var validSortBy = map[string]bool{
	"updated_at": true, "applied_date": true, "company_name": true, "status": true,
}

func (r *ApplicationRepository) List(ctx context.Context, userID int64, f entity.ApplicationListFilter) ([]entity.Application, int64, error) {
	exec := database.GetDBTx(ctx, r.db)

	args := []interface{}{userID}
	n := 2
	where := []string{"user_id = $1", "deleted_at IS NULL"}

	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		where = append(where, fmt.Sprintf("(company_name ILIKE $%d OR position_title ILIKE $%d)", n, n+1))
		args = append(args, like, like)
		n += 2
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", n))
		args = append(args, f.Status)
		n++
	}
	if f.Source != "" {
		where = append(where, fmt.Sprintf("source = $%d", n))
		args = append(args, f.Source)
		n++
	}
	if f.WorkArrangement != "" {
		where = append(where, fmt.Sprintf("work_arrangement = $%d", n))
		args = append(args, f.WorkArrangement)
		n++
	}
	if f.EmploymentType != "" {
		where = append(where, fmt.Sprintf("employment_type = $%d", n))
		args = append(args, f.EmploymentType)
		n++
	}
	if f.FromDate != "" {
		where = append(where, fmt.Sprintf("applied_date >= $%d::date", n))
		args = append(args, f.FromDate)
		n++
	}
	if f.ToDate != "" {
		where = append(where, fmt.Sprintf("applied_date <= $%d::date", n))
		args = append(args, f.ToDate)
		n++
	}
	if f.IsArchived != nil {
		where = append(where, fmt.Sprintf("is_archived = $%d", n))
		args = append(args, *f.IsArchived)
		n++
	} else {
		where = append(where, "is_archived = false")
	}

	sortBy := f.SortBy
	if !validSortBy[sortBy] {
		sortBy = "updated_at"
	}
	dir := strings.ToUpper(f.SortDir)
	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}

	var orderExpr string
	if sortBy == "applied_date" {
		sentinel := "1970-01-01"
		if dir == "ASC" {
			sentinel = "9999-12-31"
		}
		orderExpr = fmt.Sprintf("COALESCE(applied_date, '%s'::date) %s, id %s", sentinel, dir, dir)
	} else {
		orderExpr = fmt.Sprintf("%s %s, id %s", sortBy, dir, dir)
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM job_applications WHERE %s", whereClause)
	if err := sqlx.GetContext(ctx, exec, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	dataQuery := fmt.Sprintf(`
		SELECT * FROM job_applications
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderExpr, n, n+1,
	)

	var apps []entity.Application
	if err := sqlx.SelectContext(ctx, exec, &apps, dataQuery, args...); err != nil {
		return nil, 0, err
	}
	return apps, total, nil
}
