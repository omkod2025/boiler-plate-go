package auth

import (
	"context"
	"errors"
	"strings"
)

// LoginInput ข้อมูลสำหรับเข้าสู่ระบบ
type LoginInput struct {
	Email    string
	Password string
}

type AuthUseCase struct {
	credentials CredentialRepository
	verifier    PasswordVerifier
	tokens      TokenIssuer
}

func NewAuthUseCase(credentials CredentialRepository, verifier PasswordVerifier, tokens TokenIssuer) *AuthUseCase {
	return &AuthUseCase{credentials: credentials, verifier: verifier, tokens: tokens}
}

// Login ตรวจสอบอีเมลและรหัสผ่าน แล้วออก access token
// คืน ErrInvalidCredentials ทั้งกรณีไม่พบอีเมลและรหัสผ่านผิด เพื่อไม่ให้เดาได้ว่าอีเมลไหนมีในระบบ
func (uc *AuthUseCase) Login(ctx context.Context, in LoginInput) (Token, error) {
	cred, err := uc.credentials.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(in.Email)))
	if errors.Is(err, ErrCredentialNotFound) {
		return Token{}, ErrInvalidCredentials
	}
	if err != nil {
		return Token{}, err
	}
	if !uc.verifier.Verify(cred.PasswordHash, in.Password) {
		return Token{}, ErrInvalidCredentials
	}
	return uc.tokens.Issue(cred.UserID, cred.Role)
}
