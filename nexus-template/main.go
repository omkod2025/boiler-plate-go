package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"__MODULE__/app"
	"github.com/omkod2025/nexus-backend/packages/nexus-kit/migration"
)

func main() {
	if len(os.Args) > 1 {
		if len(os.Args) != 2 || os.Args[1] != "migrate" { slog.Error("use service or service migrate (up only)"); os.Exit(2) }
		ctx,cancel:=context.WithTimeout(context.Background(),10*time.Minute); defer cancel()
		if err:=migration.Up(ctx,os.Getenv("MIGRATION_DB_URL"),os.Getenv("MIGRATION_SCHEMA"),"migrations");err!=nil { slog.Error("migration stopped", "error", err);os.Exit(1) };return
	}
	if err := app.Run(context.Background()); err != nil {
		slog.Error("service stopped", "error", err)
		os.Exit(1)
	}
}
