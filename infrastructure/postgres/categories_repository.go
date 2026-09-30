package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/omkod2025/boiler-plate-go/domain/categories"
	pkgsql "github.com/omkod2025/boiler-plate-go/pkg/sql"
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
	rows, err := r.db.Query(ctx, `SELECT `+categoryColumns+` FROM public.oktf_category_get($1)`, userProfileID)
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
	err := r.db.QueryRowWithContext(ctx, `CALL public.oktp_category_insert($1, $2, $3, $4, $5, NULL, NULL, NULL)`,
		c.Name, c.UserProfileID, c.Color, c.Icon, string(c.Type)).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return categories.Category{}, err
	}
	return c, nil
}

func (r *CategoryRepository) Update(ctx context.Context, userProfileID, id int, in categories.UpdateInput) (categories.Category, error) {
	var categoryType *string
	if in.Type != nil {
		t := string(*in.Type)
		categoryType = &t
	}
	// ไม่พบ หรือเป็นของผู้ใช้อื่น → procedure คืน OUT ทุกตัวเป็น NULL
	var row categoryRow
	err := r.db.QueryRowWithContext(ctx,
		`CALL public.oktp_category_update($1, $2, $3, $4, $5, $6, NULL, NULL, NULL, NULL, NULL, NULL, NULL)`,
		id, userProfileID, in.Name, in.Color, in.Icon, categoryType).
		Scan(&row.CategoryID, &row.CategoryName, &row.Color, &row.Icon, &row.CategoryType, &row.CreatedAt, &row.UpdatedAt)
	if err != nil {
		return categories.Category{}, err
	}
	if !row.CategoryID.Valid {
		return categories.Category{}, categories.ErrNotFound
	}
	return categories.Category{
		ID: int(row.CategoryID.Int64), UserProfileID: userProfileID, Name: row.CategoryName.String,
		Color: row.Color.String, Icon: row.Icon.String, Type: categories.Type(row.CategoryType.String),
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, userProfileID, id int) error {
	var deleted bool
	if err := r.db.QueryRowWithContext(ctx, `CALL public.oktp_category_delete($1, $2, NULL)`, id, userProfileID).Scan(&deleted); err != nil {
		return err
	}
	if !deleted {
		return categories.ErrNotFound
	}
	return nil
}
