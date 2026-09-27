package users

import (
	"context"
	"errors"
)

var (
	// ErrNotFound ไม่พบผู้ใช้
	ErrNotFound = errors.New("user not found")
	// ErrEmailAlreadyExists อีเมลนี้ถูกใช้สมัครแล้ว
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// Repository port สำหรับจัดเก็บผู้ใช้ — implementation อยู่ที่ infrastructure layer
// Create ต้องคืน ErrEmailAlreadyExists เมื่ออีเมลซ้ำ และ GetByID ต้องคืน ErrNotFound เมื่อไม่พบ
type Repository interface {
	GetByID(ctx context.Context, id int) (User, error)
	Create(ctx context.Context, u User) (User, error)
}

// PasswordHasher port สำหรับ hash รหัสผ่านก่อนจัดเก็บ
type PasswordHasher interface {
	Hash(password string) (string, error)
}
