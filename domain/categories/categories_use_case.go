package categories

import (
	"context"
)

// CategoryRepositoryInterface defines the contract for category repository
type CategoryRepositoryInterface interface {
	List(ctx context.Context, userProfileID int) ([]CategoryDB, error)
	Create(ctx context.Context, in CreateCategoryDBInput) (CategoryDB, error)
	Update(ctx context.Context, id string, in UpdateCategoryDBInput) (CategoryDB, error)
	Delete(ctx context.Context, id string) error
	ListMasterCategories(ctx context.Context) ([]MasterCategoryDB, error)
	CreateMasterCategory(ctx context.Context, in CreateMasterCategoryDBInput) (MasterCategoryDB, error)
}

type CategoryUseCase struct {
	repo CategoryRepositoryInterface
}

func NewCategoryUseCase(repo CategoryRepositoryInterface) *CategoryUseCase {
	return &CategoryUseCase{repo: repo}
}

func mapDBToResponse(db CategoryDB) CategoryResponse {
	return CategoryResponse{
		ID:        int(db.CategoryID.Int64),
		Name:      db.CategoryName.String,
		Color:     db.Color.String,
		Icon:      db.Icon.String,
		Type:      db.CategoryType.String,
		CreatedAt: db.CreatedAt.String,
		UpdatedAt: db.UpdatedAt.String,
	}
}

func (uc *CategoryUseCase) List(ctx context.Context, userProfileID int) ([]CategoryResponse, error) {
	dbs, err := uc.repo.List(ctx, userProfileID)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryResponse, 0, len(dbs))
	for _, d := range dbs {
		out = append(out, mapDBToResponse(d))
	}
	return out, nil
}

func (uc *CategoryUseCase) Create(ctx context.Context, dto CreateCategoryRequest, userProfileID int) (CategoryResponse, error) {
	in := CreateCategoryDBInput{
		CategoryName:  dto.Name,
		UserProfileID: userProfileID,
		Color:         &dto.Color,
		Icon:          &dto.Icon,
		CategoryType:  dto.Type,
	}
	db, err := uc.repo.Create(ctx, in)
	if err != nil {
		return CategoryResponse{}, err
	}
	return mapDBToResponse(db), nil
}

func (uc *CategoryUseCase) Update(ctx context.Context, id string, dto UpdateCategoryRequest) (CategoryResponse, error) {
	in := UpdateCategoryDBInput{CategoryName: dto.Name, Color: dto.Color, Icon: dto.Icon, CategoryType: dto.Type}
	db, err := uc.repo.Update(ctx, id, in)
	if err != nil {
		return CategoryResponse{}, err
	}
	return mapDBToResponse(db), nil
}

func (uc *CategoryUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}
