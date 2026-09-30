package configs

import (
	"context"
	"errors"

	"github.com/omkod2025/boiler-plate-go/pkg/sql"
)

// InitPGXConnection เปิด pool และ ping ครั้งแรก — error ไม่มี DSN/รหัสผ่านติดไปด้วย
func InitPGXConnection(ctx context.Context, sqlConfig sql.PGXConfig) (*sql.PGX, error) {
	pgxConn, err := sql.NewPGX(ctx, sqlConfig)
	if err != nil {
		return nil, errors.New("database: cannot create connection pool")
	}
	if err := pgxConn.Ping(); err != nil {
		pgxConn.Close()
		return nil, errors.New("database: ping failed")
	}
	return pgxConn, nil
}
