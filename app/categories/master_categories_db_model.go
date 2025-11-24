package categories

import "time"

// Master Categories Database model aligned with okdt_master_categories table
type MasterCategoryDB struct {
	CategoryID   int       `db:"category_id"`
	CategoryName string    `db:"category_name"`
	Color        string    `db:"color"`
	Icon         string    `db:"icon"`
	CategoryType string    `db:"category_type"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// DB input payloads for master categories
type CreateMasterCategoryDBInput struct {
	CategoryName string
	Color        string
	Icon         string
	CategoryType string
}

type UpdateMasterCategoryDBInput struct {
	CategoryName *string
	Color        *string
	Icon         *string
	CategoryType *string
}
