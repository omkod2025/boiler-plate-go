package categories

import (
	"context"
	"errors"
)

var (
	// ErrNotFound ไม่พบหมวดหมู่ หรือหมวดหมู่ไม่ใช่ของผู้ใช้คนนี้
	ErrNotFound = errors.New("category not found")
	// ErrInvalidType ประเภทหมวดหมู่ไม่ถูกต้อง
	ErrInvalidType = errors.New("invalid category type")
)

// UpdateInput ข้อมูลที่ต้องการแก้ไข — field ที่เป็น nil จะไม่ถูกแก้
type UpdateInput struct {
	Name  *string
	Color *string
	Icon  *string
	Type  *Type
}

// Repository port สำหรับจัดเก็บหมวดหมู่ — implementation อยู่ที่ infrastructure layer
// Update และ Delete ต้องคืน ErrNotFound เมื่อไม่พบ id หรือ id ไม่ใช่ของ userProfileID
type Repository interface {
	ListByUser(ctx context.Context, userProfileID int) ([]Category, error)
	Create(ctx context.Context, c Category) (Category, error)
	Update(ctx context.Context, userProfileID, id int, in UpdateInput) (Category, error)
	Delete(ctx context.Context, userProfileID, id int) error
}
