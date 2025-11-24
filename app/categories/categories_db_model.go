package categories

import (
	"database/sql"
)

// Database model aligned with migration columns
type CategoryDB struct {
	CategoryID    sql.NullInt64
	CategoryName  sql.NullString
	UserProfileID sql.NullInt64
	Color         sql.NullString
	Icon          sql.NullString
	CategoryType  sql.NullString
	CreatedAt     sql.NullString
	UpdatedAt     sql.NullString
}

// DB input payloads used strictly by repository layer
type CreateCategoryDBInput struct {
	CategoryName  string
	UserProfileID int
	Color         *string
	Icon          *string
	CategoryType  string
}

type UpdateCategoryDBInput struct {
	CategoryName *string
	Color        *string
	Icon         *string
	CategoryType *string
}
