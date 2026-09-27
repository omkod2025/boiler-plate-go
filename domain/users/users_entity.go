package users

import "time"

// Role สิทธิ์ของผู้ใช้
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// User entity ของผู้ใช้ในระบบ
type User struct {
	ID           int
	Email        string
	Name         string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
