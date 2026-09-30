package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/omkod2025/boiler-plate-go/domain/auth"
	pkgsql "github.com/omkod2025/boiler-plate-go/pkg/sql"
)

// CredentialRepository implements auth.CredentialRepository โดยอ่านจากตารางผู้ใช้
type CredentialRepository struct {
	db *pkgsql.PGX
}

var _ auth.CredentialRepository = (*CredentialRepository)(nil)

func NewCredentialRepository(db *pkgsql.PGX) *CredentialRepository {
	return &CredentialRepository{db: db}
}

func (r *CredentialRepository) FindByEmail(ctx context.Context, email string) (auth.Credential, error) {
	var c auth.Credential
	err := r.db.QueryRowWithContext(ctx,
		`SELECT user_profile_id, password_hash, role FROM public.oktf_user_credential_get($1)`, email).
		Scan(&c.UserID, &c.PasswordHash, &c.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Credential{}, auth.ErrCredentialNotFound
	}
	return c, err
}
