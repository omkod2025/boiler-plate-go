package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/omkod2025/boiler-plate-go/domain/categories"
	"github.com/omkod2025/boiler-plate-go/domain/users"
	"github.com/omkod2025/boiler-plate-go/pkg/testhelper"
)

// Integration test กับ PostgreSQL จริง (migration ใน migrations/) — ตัวอย่างการใช้ testhelper
func TestRepositoriesAgainstPostgres(t *testing.T) {
	db := testhelper.Postgres(t)
	ctx := context.Background()
	userRepo := NewUserRepository(db)
	catRepo := NewCategoryRepository(db)

	alice, err := userRepo.Create(ctx, users.User{Email: "alice@example.com", Name: "Alice", PasswordHash: "h", Role: users.Role("user")})
	if err != nil {
		t.Fatal(err)
	}
	bob, err := userRepo.Create(ctx, users.User{Email: "bob@example.com", Name: "Bob", PasswordHash: "h", Role: users.Role("user")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := userRepo.Create(ctx, users.User{Email: "alice@example.com", Name: "Again", PasswordHash: "h", Role: users.Role("user")}); err == nil {
		t.Fatal("duplicate email must fail")
	}

	food, err := catRepo.Create(ctx, categories.Category{UserProfileID: alice.ID, Name: "Food", Color: "#f00", Icon: "noodle", Type: categories.TypeExpense})
	if err != nil {
		t.Fatal(err)
	}
	list, err := catRepo.ListByUser(ctx, alice.ID)
	if err != nil || len(list) != 1 || list[0].Name != "Food" {
		t.Fatalf("list %v %v", list, err)
	}
	if other, _ := catRepo.ListByUser(ctx, bob.ID); len(other) != 0 {
		t.Fatal("bob must not see alice's categories")
	}
	name := "Meals"
	updated, err := catRepo.Update(ctx, alice.ID, food.ID, categories.UpdateInput{Name: &name})
	if err != nil || updated.Name != "Meals" || updated.Color != "#f00" {
		t.Fatalf("update %v %v", updated, err)
	}
	if _, err := catRepo.Update(ctx, bob.ID, food.ID, categories.UpdateInput{Name: &name}); !errors.Is(err, categories.ErrNotFound) {
		t.Fatalf("update by another user: %v", err)
	}
	if err := catRepo.Delete(ctx, bob.ID, food.ID); !errors.Is(err, categories.ErrNotFound) {
		t.Fatalf("delete by another user: %v", err)
	}
	if err := catRepo.Delete(ctx, alice.ID, food.ID); err != nil {
		t.Fatal(err)
	}
}
