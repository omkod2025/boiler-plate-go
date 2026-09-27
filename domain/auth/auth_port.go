package auth

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrInvalidCredentials อีเมลหรือรหัสผ่านไม่ถูกต้อง
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrCredentialNotFound ไม่พบข้อมูลสำหรับเข้าสู่ระบบ — ใช้ระหว่าง repository กับ use case เท่านั้น
	ErrCredentialNotFound = errors.New("credential not found")
)

// Credential ข้อมูลที่ใช้ตรวจสอบการเข้าสู่ระบบ
type Credential struct {
	UserID       int
	PasswordHash string
	Role         string
}

// Token ผลลัพธ์ของการเข้าสู่ระบบ
type Token struct {
	AccessToken string
	ExpiresIn   time.Duration
}

// CredentialRepository port สำหรับค้นหา credential — ต้องคืน ErrCredentialNotFound เมื่อไม่พบ
type CredentialRepository interface {
	FindByEmail(ctx context.Context, email string) (Credential, error)
}

// PasswordVerifier port สำหรับตรวจรหัสผ่านกับ hash
type PasswordVerifier interface {
	Verify(hash, password string) bool
}

// TokenIssuer port สำหรับออก access token
type TokenIssuer interface {
	Issue(userID int, role string) (Token, error)
}
