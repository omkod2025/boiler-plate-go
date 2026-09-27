package categories

import (
	"context"
	"errors"
	"testing"
)

type fakeRepo struct {
	created   *Category
	updatedBy int
	deletedBy int
}

func (f *fakeRepo) ListByUser(ctx context.Context, userProfileID int) ([]Category, error) {
	return nil, nil
}

func (f *fakeRepo) Create(ctx context.Context, c Category) (Category, error) {
	f.created = &c
	return c, nil
}

func (f *fakeRepo) Update(ctx context.Context, userProfileID, id int, in UpdateInput) (Category, error) {
	f.updatedBy = userProfileID
	return Category{ID: id, UserProfileID: userProfileID}, nil
}

func (f *fakeRepo) Delete(ctx context.Context, userProfileID, id int) error {
	f.deletedBy = userProfileID
	return nil
}

func TestCreateSetsOwner(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewCategoryUseCase(repo)

	_, err := uc.Create(context.Background(), 7, CreateInput{Name: "Food", Type: TypeExpense})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.created == nil || repo.created.UserProfileID != 7 {
		t.Fatalf("expected category owned by user 7, got %+v", repo.created)
	}
}

func TestCreateRejectsInvalidType(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewCategoryUseCase(repo)

	_, err := uc.Create(context.Background(), 7, CreateInput{Name: "Food", Type: "other"})
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
	if repo.created != nil {
		t.Fatal("repository must not be called for invalid input")
	}
}

func TestUpdateRejectsInvalidType(t *testing.T) {
	uc := NewCategoryUseCase(&fakeRepo{})
	bad := Type("other")

	_, err := uc.Update(context.Background(), 7, 1, UpdateInput{Type: &bad})
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}

func TestUpdateAndDeleteScopedToUser(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewCategoryUseCase(repo)

	if _, err := uc.Update(context.Background(), 7, 1, UpdateInput{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := uc.Delete(context.Background(), 7, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updatedBy != 7 || repo.deletedBy != 7 {
		t.Fatalf("expected repository calls scoped to user 7, got update=%d delete=%d", repo.updatedBy, repo.deletedBy)
	}
}
