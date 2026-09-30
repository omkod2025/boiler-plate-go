package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/omkod2025/boiler-plate-go/domain/users"
	pkgsql "github.com/omkod2025/boiler-plate-go/pkg/sql"
)

// Repository ใน package นี้ไม่อ่าน/เขียนตารางตรง: อ่านผ่าน function oktf_* ด้วย SELECT และเขียนผ่าน
// procedure oktp_* ด้วย CALL (ผลลัพธ์มาจาก OUT parameters) — schema และ routine อยู่ใน migrations/

const pgUniqueViolation = "23505"

func scanUser(s rowScanner) (users.User, error) {
	var u users.User
	var role string
	if err := s.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &role, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return users.User{}, users.ErrNotFound
		}
		return users.User{}, err
	}
	u.Role = users.Role(role)
	return u, nil
}

// UserRepository implements users.Repository ด้วย PostgreSQL
type UserRepository struct {
	db *pkgsql.PGX
}

var _ users.Repository = (*UserRepository)(nil)

func NewUserRepository(db *pkgsql.PGX) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (users.User, error) {
	return scanUser(r.db.QueryRowWithContext(ctx,
		`SELECT user_profile_id, email, full_name, password_hash, role, created_at, updated_at FROM public.oktf_user_get($1)`, id))
}

func (r *UserRepository) Create(ctx context.Context, u users.User) (users.User, error) {
	err := r.db.QueryRowWithContext(ctx, `CALL public.oktp_user_insert($1, $2, $3, $4, NULL, NULL, NULL)`,
		u.Email, u.Name, u.PasswordHash, string(u.Role)).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return users.User{}, users.ErrEmailAlreadyExists
	}
	if err != nil {
		return users.User{}, err
	}
	return u, nil
}
