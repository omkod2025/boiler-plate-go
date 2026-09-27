package categories

import (
	"context"
)

// CreateInput ข้อมูลสำหรับสร้างหมวดหมู่
type CreateInput struct {
	Name  string
	Color string
	Icon  string
	Type  Type
}

type CategoryUseCase struct {
	repo Repository
}

func NewCategoryUseCase(repo Repository) *CategoryUseCase {
	return &CategoryUseCase{repo: repo}
}

func (uc *CategoryUseCase) List(ctx context.Context, userProfileID int) ([]Category, error) {
	return uc.repo.ListByUser(ctx, userProfileID)
}

func (uc *CategoryUseCase) Create(ctx context.Context, userProfileID int, in CreateInput) (Category, error) {
	if !in.Type.Valid() {
		return Category{}, ErrInvalidType
	}
	return uc.repo.Create(ctx, Category{
		UserProfileID: userProfileID,
		Name:          in.Name,
		Color:         in.Color,
		Icon:          in.Icon,
		Type:          in.Type,
	})
}

// Update แก้ไขหมวดหมู่ได้เฉพาะของ userProfileID เท่านั้น
func (uc *CategoryUseCase) Update(ctx context.Context, userProfileID, id int, in UpdateInput) (Category, error) {
	if in.Type != nil && !in.Type.Valid() {
		return Category{}, ErrInvalidType
	}
	return uc.repo.Update(ctx, userProfileID, id, in)
}

// Delete ลบหมวดหมู่ได้เฉพาะของ userProfileID เท่านั้น
func (uc *CategoryUseCase) Delete(ctx context.Context, userProfileID, id int) error {
	return uc.repo.Delete(ctx, userProfileID, id)
}
