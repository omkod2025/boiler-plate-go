package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/omkod2025/boiler-plate-go/app"
	"github.com/omkod2025/boiler-plate-go/configs"
	"github.com/omkod2025/boiler-plate-go/pkg/logger"
	"github.com/omkod2025/boiler-plate-go/pkg/telemetry"
)

func main() {
	os.Exit(run())
}

// run คืน exit code — แยกจาก main เพื่อให้ defer (flush trace) ทำงานก่อน os.Exit
func run() int {
	// role เลือกจาก flag -role หรือ APP_ROLE (ค่าเริ่มต้น api) — image เดียวรันได้ทุก role
	defaultRole := os.Getenv("APP_ROLE")
	if defaultRole == "" {
		defaultRole = app.RoleAPI
	}
	role := flag.String("role", defaultRole, "api | stream | worker | receiver | sandbox | migrate")
	flag.Parse()

	// SIGINT/SIGTERM cancel ctx → ทุก role หยุดรับงานใหม่แล้วรองานที่ค้างให้จบ
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := configs.LoadConfig(*role)

	shutdownTracing, err := telemetry.Init(ctx, telemetry.Options{
		ServiceName: cfg.Env.APP_NAME, Version: cfg.Env.APP_VERSION, Environment: cfg.Env.APP_ENV,
		Role: cfg.Env.APP_ROLE, Endpoint: cfg.Env.OTEL_ENDPOINT,
	})
	if err != nil {
		logger.Error("telemetry: ", err)
		return 1
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdownTracing(flushCtx)
	}()

	if err := app.Run(ctx, cfg); err != nil {
		logger.Error("role "+cfg.Env.APP_ROLE+" failed: ", err)
		return 1
	}
	logger.Info("role " + cfg.Env.APP_ROLE + " stopped")
	return 0
}
