package auth

import (
	"context"
	"errors"
	"testing"
)

type fakeCredentials struct {
	byEmail  map[string]Credential
	gotEmail string
}

func (f *fakeCredentials) FindByEmail(ctx context.Context, email string) (Credential, error) {
	f.gotEmail = email
	c, ok := f.byEmail[email]
	if !ok {
		return Credential{}, ErrCredentialNotFound
	}
	return c, nil
}

type fakeVerifier struct{}

func (fakeVerifier) Verify(hash, password string) bool { return hash == "hashed:"+password }

type fakeIssuer struct{}

func (fakeIssuer) Issue(userID int, role string) (Token, error) {
	return Token{AccessToken: role + "-token"}, nil
}

func newTestUseCase() (*AuthUseCase, *fakeCredentials) {
	creds := &fakeCredentials{byEmail: map[string]Credential{
		"foo@example.com": {UserID: 1, PasswordHash: "hashed:secret123", Role: "user"},
	}}
	return NewAuthUseCase(creds, fakeVerifier{}, fakeIssuer{}), creds
}

func TestLoginSuccess(t *testing.T) {
	uc, creds := newTestUseCase()

	token, err := uc.Login(context.Background(), LoginInput{Email: " Foo@Example.com", Password: "secret123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.gotEmail != "foo@example.com" {
		t.Fatalf("expected normalized email lookup, got %q", creds.gotEmail)
	}
	if token.AccessToken != "user-token" {
		t.Fatalf("unexpected token %q", token.AccessToken)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	uc, _ := newTestUseCase()

	cases := []LoginInput{
		{Email: "foo@example.com", Password: "wrong"},
		{Email: "nobody@example.com", Password: "secret123"},
	}
	for _, in := range cases {
		if _, err := uc.Login(context.Background(), in); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("login %q: expected ErrInvalidCredentials, got %v", in.Email, err)
		}
	}
}
