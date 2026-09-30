package configs

import (
	"context"
	"github.com/omkod2025/boiler-plate-go/pkg/sql"
)

func InitPGXConnection(ctx context.Context, sqlConfig sql.PGXConfig) (*sql.PGX, error) {
	pgxConn, err := sql.NewPGX(ctx, sqlConfig)
	if err != nil {
		panic(err)
	}
	return pgxConn, nil
}
