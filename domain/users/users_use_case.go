package users

import (
	"context"
	"strings"
)

// RegisterInput ข้อมูลสำหรับสมัครสมาชิก
type RegisterInput struct {
	Email    string
	Name     string
	Password string
}

type UserUseCase struct {
	repo   Repository
	hasher PasswordHasher
}

func NewUserUseCase(repo Repository, hasher PasswordHasher) *UserUseCase {
	return &UserUseCase{repo: repo, hasher: hasher}
}

// Register สร้างผู้ใช้ใหม่ด้วย role เริ่มต้นเป็น RoleUser
func (uc *UserUseCase) Register(ctx context.Context, in RegisterInput) (User, error) {
	hash, err := uc.hasher.Hash(in.Password)
	if err != nil {
		return User{}, err
	}
	return uc.repo.Create(ctx, User{
		Email:        strings.ToLower(strings.TrimSpace(in.Email)),
		Name:         strings.TrimSpace(in.Name),
		PasswordHash: hash,
		Role:         RoleUser,
	})
}

// GetProfile ดึงข้อมูลผู้ใช้ตาม id
func (uc *UserUseCase) GetProfile(ctx context.Context, id int) (User, error) {
	return uc.repo.GetByID(ctx, id)
}
