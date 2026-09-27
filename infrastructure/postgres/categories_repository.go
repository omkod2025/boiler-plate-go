package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/omkod2025-boop/omgon-notification-service/domain/categories"
	pkgsql "github.com/omkod2025-boop/omgon-notification-service/pkg/sql"
)

const categoryColumns = `category_id, category_name, user_profile_id, color, icon, category_type, created_at, updated_at`

// categoryRow database model aligned with okdt_categories columns
type categoryRow struct {
	CategoryID    sql.NullInt64
	CategoryName  sql.NullString
	UserProfileID sql.NullInt64
	Color         sql.NullString
	Icon          sql.NullString
	CategoryType  sql.NullString
	CreatedAt     sql.NullTime
	UpdatedAt     sql.NullTime
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCategory(s rowScanner) (categories.Category, error) {
	var r categoryRow
	if err := s.Scan(&r.CategoryID, &r.CategoryName, &r.UserProfileID, &r.Color, &r.Icon, &r.CategoryType, &r.CreatedAt, &r.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return categories.Category{}, categories.ErrNotFound
		}
		return categories.Category{}, err
	}
	return categories.Category{
		ID:            int(r.CategoryID.Int64),
		UserProfileID: int(r.UserProfileID.Int64),
		Name:          r.CategoryName.String,
		Color:         r.Color.String,
		Icon:          r.Icon.String,
		Type:          categories.Type(r.CategoryType.String),
		CreatedAt:     r.CreatedAt.Time,
		UpdatedAt:     r.UpdatedAt.Time,
	}, nil
}

// CategoryRepository implements categories.Repository ด้วย PostgreSQL
type CategoryRepository struct {
	db *pkgsql.PGX
}

var _ categories.Repository = (*CategoryRepository)(nil)

func NewCategoryRepository(db *pkgsql.PGX) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) ListByUser(ctx context.Context, userProfileID int) ([]categories.Category, error) {
	fncName := "repo categories ListByUser"
	rows, err := r.db.Query(ctx, `SELECT `+categoryColumns+` FROM oktf_category_get($1);`, userProfileID)
	if err != nil {
		slog.Error(fncName, "err", err)
		return nil, err
	}
	defer rows.Close()
	result := []categories.Category{}
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			slog.Error(fncName, "row error", err)
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *CategoryRepository) Create(ctx context.Context, c categories.Category) (categories.Category, error) {
	return scanCategory(r.db.QueryRowWithContext(ctx, `
        INSERT INTO okdt_categories(category_name, user_profile_id, color, icon, category_type)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING `+categoryColumns,
		c.Name, c.UserProfileID, c.Color, c.Icon, string(c.Type)))
}

func (r *CategoryRepository) Update(ctx context.Context, userProfileID, id int, in categories.UpdateInput) (categories.Category, error) {
	var categoryType *string
	if in.Type != nil {
		t := string(*in.Type)
		categoryType = &t
	}
	return scanCategory(r.db.QueryRowWithContext(ctx, `
        UPDATE okdt_categories SET
          category_name = COALESCE($3, category_name),
          color = COALESCE($4, color),
          icon = COALESCE($5, icon),
          category_type = COALESCE($6, category_type),
          updated_at = now()
        WHERE category_id = $1 AND user_profile_id = $2
        RETURNING `+categoryColumns,
		id, userProfileID, in.Name, in.Color, in.Icon, categoryType))
}

func (r *CategoryRepository) Delete(ctx context.Context, userProfileID, id int) error {
	var deletedID int
	err := r.db.QueryRowWithContext(ctx,
		`DELETE FROM okdt_categories WHERE category_id = $1 AND user_profile_id = $2 RETURNING category_id`,
		id, userProfileID).Scan(&deletedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return categories.ErrNotFound
	}
	return err
}
