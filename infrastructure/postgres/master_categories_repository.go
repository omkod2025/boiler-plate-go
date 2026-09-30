package postgres

import (
	"context"
	"time"

	"github.com/omkod2025/boiler-plate-go/domain/categories"
	pkgsql "github.com/omkod2025/boiler-plate-go/pkg/sql"
)

// masterCategoryRow database model aligned with okdt_master_categories table
type masterCategoryRow struct {
	CategoryID   int
	CategoryName string
	Color        string
	Icon         string
	CategoryType string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (r masterCategoryRow) toEntity() categories.MasterCategory {
	return categories.MasterCategory{
		ID:        r.CategoryID,
		Name:      r.CategoryName,
		Color:     r.Color,
		Icon:      r.Icon,
		Type:      categories.Type(r.CategoryType),
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// MasterCategoryRepository จัดการหมวดหมู่ตั้งต้นของระบบ
// ยังไม่มี use case เรียกใช้ — ให้ประกาศ port ใน domain/categories เมื่อมี use case ที่ต้องใช้
type MasterCategoryRepository struct {
	db *pkgsql.PGX
}

func NewMasterCategoryRepository(db *pkgsql.PGX) *MasterCategoryRepository {
	return &MasterCategoryRepository{db: db}
}

func (r *MasterCategoryRepository) List(ctx context.Context) ([]categories.MasterCategory, error) {
	rows, err := r.db.Query(ctx, `SELECT category_id, category_name, color, icon, category_type, created_at, updated_at
      FROM public.oktf_master_category_list()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []categories.MasterCategory{}
	for rows.Next() {
		var c masterCategoryRow
		if err := rows.Scan(&c.CategoryID, &c.CategoryName, &c.Color, &c.Icon, &c.CategoryType, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, c.toEntity())
	}
	return result, rows.Err()
}

func (r *MasterCategoryRepository) Create(ctx context.Context, c categories.MasterCategory) (categories.MasterCategory, error) {
	err := r.db.QueryRowWithContext(ctx, `CALL public.oktp_master_category_insert($1, $2, $3, $4, NULL, NULL, NULL)`,
		c.Name, c.Color, c.Icon, string(c.Type)).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}
