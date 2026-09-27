package users

import (
	"context"
	"testing"
)

type fakeRepo struct {
	created *User
}

func (f *fakeRepo) GetByID(ctx context.Context, id int) (User, error) {
	return User{}, ErrNotFound
}

func (f *fakeRepo) Create(ctx context.Context, u User) (User, error) {
	f.created = &u
	return u, nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }

func TestRegisterHashesPasswordAndNormalizesEmail(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewUserUseCase(repo, fakeHasher{})

	_, err := uc.Register(context.Background(), RegisterInput{Email: "  Foo@Example.COM ", Name: " Foo ", Password: "secret123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := repo.created
	if got == nil {
		t.Fatal("expected repository Create to be called")
	}
	if got.Email != "foo@example.com" || got.Name != "Foo" {
		t.Fatalf("expected normalized email/name, got %q %q", got.Email, got.Name)
	}
	if got.PasswordHash != "hashed:secret123" {
		t.Fatalf("expected hashed password, got %q", got.PasswordHash)
	}
	if got.Role != RoleUser {
		t.Fatalf("expected default role %q, got %q", RoleUser, got.Role)
	}
}
