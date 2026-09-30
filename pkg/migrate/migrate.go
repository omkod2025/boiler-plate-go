// Package migrate รัน goose migration ที่ฝังใน binary (migrations.FS) กับ pgx pool
package migrate

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Up รัน migration ทั้งหมดที่ยังไม่ได้รันจาก fsys (ไฟล์ *.sql ที่ root) และคืน version ปัจจุบัน
// goose ล็อกตาราง goose_db_version ไว้ระหว่างรัน จึงรันพร้อมกันหลาย instance ได้ปลอดภัย
func Up(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) (int64, error) {
	db := stdlib.OpenDBFromPool(pool)
	defer func() { _ = db.Close() }()
	provider, err := goose.NewProvider(goose.DialectPostgres, db, fsys)
	if err != nil {
		return 0, fmt.Errorf("migrate: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return 0, fmt.Errorf("migrate: up: %w", err)
	}
	return provider.GetDBVersion(ctx)
}
