package handler

import (
	"time"

	"github.com/omkod2025-boop/omgon-notification-service/domain/categories"
)

// Request DTOs
type CreateCategoryRequest struct {
	Name  string `json:"name" th:"ชื่อหมวดหมู่" binding:"required,min=1,max=100,no_sql_inject"`
	Color string `json:"color" th:"สี" binding:"required,max=20"`
	Icon  string `json:"icon" th:"ไอคอน" binding:"required,max=16"`
	Type  string `json:"type" th:"ประเภท" binding:"required,oneof=income expense both"`
}

func (r CreateCategoryRequest) toInput() categories.CreateInput {
	return categories.CreateInput{
		Name:  r.Name,
		Color: r.Color,
		Icon:  r.Icon,
		Type:  categories.Type(r.Type),
	}
}

type UpdateCategoryRequest struct {
	Name  *string `json:"name,omitempty" th:"ชื่อหมวดหมู่" binding:"omitempty,max=100,no_sql_inject"`
	Color *string `json:"color,omitempty" th:"สี" binding:"omitempty,max=20,no_sql_inject"`
	Icon  *string `json:"icon,omitempty" th:"ไอคอน" binding:"omitempty,max=16"`
	Type  *string `json:"type,omitempty" th:"ประเภท" binding:"omitempty,oneof=income expense both"`
}

func (r UpdateCategoryRequest) toInput() categories.UpdateInput {
	in := categories.UpdateInput{Name: r.Name, Color: r.Color, Icon: r.Icon}
	if r.Type != nil {
		t := categories.Type(*r.Type)
		in.Type = &t
	}
	return in
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

func toCategoryResponse(c categories.Category) CategoryResponse {
	return CategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Color:     c.Color,
		Icon:      c.Icon,
		Type:      string(c.Type),
		CreatedAt: formatTime(c.CreatedAt),
		UpdatedAt: formatTime(c.UpdatedAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
