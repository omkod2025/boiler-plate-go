package categories

// Request DTOs
type CreateCategoryRequest struct {
	Name  string `json:"name" th:"ชื่อหมวดหมู่" binding:"required,min=1,max=100,no_sql_inject"`
	Color string `json:"color" th:"สี" binding:"required,max=20"`
	Icon  string `json:"icon" th:"ไอคอน" binding:"required,max=16"`
	Type  string `json:"type" th:"ประเภท" binding:"required,oneof=income expense both"`
}

type UpdateCategoryRequest struct {
	Name  *string `json:"name,omitempty" th:"ชื่อหมวดหมู่" binding:"omitempty,max=100,no_sql_inject"`
	Color *string `json:"color,omitempty" th:"สี" binding:"omitempty,max=20,no_sql_inject"`
	Icon  *string `json:"icon,omitempty" th:"ไอคอน" binding:"omitempty,max=16"`
	Type  *string `json:"type,omitempty" th:"ประเภท" binding:"omitempty,oneof=income expense both"`
}

// Response DTO
type CategoryResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	Icon      string `json:"icon"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
