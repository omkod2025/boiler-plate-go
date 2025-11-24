package categories

import (
	"context"
	"log/slog"
	pkgsql "github.com/omkod2025-boop/omgon-notification-service/pkg/sql"
)

type CategoryRepository struct {
	db *pkgsql.PGX
}

func NewCategoryRepository(db *pkgsql.PGX) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List(ctx context.Context, userProfileID int) ([]CategoryDB, error) {
	slog.Info("", "userProfileID", userProfileID)
	fncName := "repo categories List"
	rows, err := r.db.Query(ctx, `SELECT 
        category_id,
        category_name,
        user_profile_id,
        color,
        icon,
        category_type,
        created_at,
        updated_at
      FROM oktf_category_get($1);`, userProfileID)
	if err != nil {
		slog.Error(fncName, "err", err)
		return nil, err
	}
	defer rows.Close()
	var result []CategoryDB
	for rows.Next() {
		var c CategoryDB
		if err := rows.Scan(&c.CategoryID, &c.CategoryName, &c.UserProfileID, &c.Color, &c.Icon, &c.CategoryType, &c.CreatedAt, &c.UpdatedAt); err != nil {
			slog.Error(fncName, "row error", fncName)
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
}

func (r *CategoryRepository) Create(ctx context.Context, in CreateCategoryDBInput) (CategoryDB, error) {
	var c CategoryDB
	err := r.db.QueryRowWithContext(ctx, `
        INSERT INTO okdt_categories(category_name, user_profile_id, color, icon, category_type)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING category_id, category_name, user_profile_id, color, icon, category_type, created_at, updated_at
    `, in.CategoryName, in.UserProfileID, in.Color, in.Icon, in.CategoryType).Scan(&c.CategoryID, &c.CategoryName, &c.UserProfileID, &c.Color, &c.Icon, &c.CategoryType, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *CategoryRepository) Update(ctx context.Context, id string, in UpdateCategoryDBInput) (CategoryDB, error) {
	var c CategoryDB
	err := r.db.QueryRowWithContext(ctx, `
        UPDATE okdt_categories SET
          color = COALESCE($2, color),
          icon = COALESCE($3, icon),
          category_type = COALESCE($4, category_type),
          updated_at = now()
        WHERE category_id = $1::int
        RETURNING category_id, category_name, user_profile_id, color, icon, category_type, created_at, updated_at
    `, id, in.Color, in.Icon, in.CategoryType).Scan(&c.CategoryID, &c.CategoryName, &c.UserProfileID, &c.Color, &c.Icon, &c.CategoryType, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *CategoryRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecWithContext(ctx, `DELETE FROM okdt_categories WHERE category_id=$1::int`, id)
	return err
}

// Master Categories Repository Methods
func (r *CategoryRepository) ListMasterCategories(ctx context.Context) ([]MasterCategoryDB, error) {
	rows, err := r.db.Query(ctx, `SELECT 
        category_id,
        category_name,
        color,
        icon,
        category_type,
        created_at,
        updated_at
      FROM okdt_master_categories 
      ORDER BY category_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []MasterCategoryDB
	for rows.Next() {
		var c MasterCategoryDB
		if err := rows.Scan(&c.CategoryID, &c.CategoryName, &c.Color, &c.Icon, &c.CategoryType, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
}

func (r *CategoryRepository) CreateMasterCategory(ctx context.Context, in CreateMasterCategoryDBInput) (MasterCategoryDB, error) {
	var c MasterCategoryDB
	err := r.db.QueryRowWithContext(ctx, `
        INSERT INTO okdt_master_categories(category_name, color, icon, category_type)
        VALUES ($1, $2, $3, $4)
        RETURNING category_id, category_name, color, icon, category_type, created_at, updated_at
    `, in.CategoryName, in.Color, in.Icon, in.CategoryType).Scan(&c.CategoryID, &c.CategoryName, &c.Color, &c.Icon, &c.CategoryType, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}
