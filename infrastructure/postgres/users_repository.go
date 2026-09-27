package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/omkod2025-boop/omgon-notification-service/domain/users"
	pkgsql "github.com/omkod2025-boop/omgon-notification-service/pkg/sql"
)

// ตัวอย่าง schema ที่ repository นี้คาดหวัง:
//
//	CREATE TABLE okdt_user_profiles (
//	  user_profile_id SERIAL PRIMARY KEY,
//	  email           TEXT NOT NULL UNIQUE,
//	  full_name       TEXT NOT NULL,
//	  password_hash   TEXT NOT NULL,
//	  role            TEXT NOT NULL DEFAULT 'user',
//	  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
//	  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
//	);
const userColumns = `user_profile_id, email, full_name, password_hash, role, created_at, updated_at`

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
		`SELECT `+userColumns+` FROM okdt_user_profiles WHERE user_profile_id = $1`, id))
}

func (r *UserRepository) Create(ctx context.Context, u users.User) (users.User, error) {
	created, err := scanUser(r.db.QueryRowWithContext(ctx, `
        INSERT INTO okdt_user_profiles(email, full_name, password_hash, role)
        VALUES ($1, $2, $3, $4)
        RETURNING `+userColumns,
		u.Email, u.Name, u.PasswordHash, string(u.Role)))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return users.User{}, users.ErrEmailAlreadyExists
	}
	return created, err
}
