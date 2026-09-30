package handler

import (
	"github.com/omkod2025/boiler-plate-go/domain/users"
)

// Request DTOs
type RegisterUserRequest struct {
	Email    string `json:"email" th:"อีเมล" binding:"required,email,max=255"`
	Name     string `json:"name" th:"ชื่อ" binding:"required,min=1,max=100,no_sql_inject"`
	Password string `json:"password" th:"รหัสผ่าน" binding:"required,min=8,max=72"`
}

func (r RegisterUserRequest) toInput() users.RegisterInput {
	return users.RegisterInput{Email: r.Email, Name: r.Name, Password: r.Password}
}

// Response DTO — ไม่ส่ง password hash ออกไป
type UserResponse struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func toUserResponse(u users.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      string(u.Role),
		CreatedAt: formatTime(u.CreatedAt),
		UpdatedAt: formatTime(u.UpdatedAt),
	}
}
