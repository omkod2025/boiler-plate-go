// Package testhelper เปิด PostgreSQL และ RabbitMQ จริงใน container (testcontainers-go) สำหรับ integration test
// ใช้เฉพาะในไฟล์ *_test.go
//
//	db := testhelper.Postgres(t)          // database ใหม่ที่รัน migration แล้ว
//	url := testhelper.RabbitMQ(t)         // amqp://guest:guest@host:port/
//
// ไม่มี Docker: test จะ skip (ตั้ง REQUIRE_DOCKER=1 ใน CI ให้ fail แทน) · `go test -short` ข้าม test เหล่านี้
package testhelper

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcrabbit "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/omkod2025/boiler-plate-go/migrations"
	"github.com/omkod2025/boiler-plate-go/pkg/migrate"
	"github.com/omkod2025/boiler-plate-go/pkg/sql"
)

// PostgresImage ตรงกับ production (PostgreSQL 18)
const PostgresImage = "postgres:18"

// RabbitImage ตรงกับ production
const RabbitImage = "rabbitmq:4.1-management"

var (
	pgOnce sync.Once
	pgCont *tcpostgres.PostgresContainer
	pgErr  error

	mqOnce sync.Once
	mqURL  string
	mqErr  error
)

// RequireDocker skip test เมื่อไม่มี Docker หรือรันด้วย -short
func RequireDocker(t testing.TB) {
	t.Helper()
	if testing.Short() {
		t.Skip("testhelper: -short skips container tests")
	}
	ok := func() (ok bool) {
		defer func() {
			if recover() != nil {
				ok = false
			}
		}()
		p, err := testcontainers.NewDockerProvider()
		if err != nil {
			return false
		}
		defer func() { _ = p.Close() }()
		return p.Health(context.Background()) == nil
	}()
	if ok {
		return
	}
	if os.Getenv("REQUIRE_DOCKER") == "1" {
		t.Fatal("testhelper: Docker is required (REQUIRE_DOCKER=1) but not available")
	}
	t.Skip("testhelper: Docker not available")
}

// Postgres คืน connection ไปยัง database ใหม่ (container ใช้ร่วมกันทั้ง test binary) ที่รัน migration แล้ว
// database ถูกลบเมื่อ test จบ
func Postgres(t testing.TB) *sql.PGX {
	t.Helper()
	RequireDocker(t)
	ctx := context.Background()
	pgOnce.Do(func() {
		pgCont, pgErr = tcpostgres.Run(ctx, PostgresImage,
			tcpostgres.WithDatabase("app"), tcpostgres.WithUsername("postgres"), tcpostgres.WithPassword("postgres"),
			testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(2*time.Minute)))
	})
	if pgErr != nil {
		t.Fatalf("testhelper: start postgres: %v", pgErr)
	}
	host, _ := pgCont.Host(ctx)
	port, _ := pgCont.MappedPort(ctx, "5432/tcp")
	admin, err := sql.NewPGX(ctx, sql.PGXConfig{DB_HOST: host, DB_PORT: port.Port(), DB_USER: "postgres", DB_PASSWORD: "postgres", DB_NAME: "app", DB_MAX_CONNS: 2, DB_MIN_CONNS: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	name := "t_" + randomSuffix()
	if _, err := admin.GetDB().Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatal(err)
	}
	db, err := sql.NewPGX(ctx, sql.PGXConfig{DB_HOST: host, DB_PORT: port.Port(), DB_USER: "postgres", DB_PASSWORD: "postgres", DB_NAME: name, DB_MAX_CONNS: 5, DB_MIN_CONNS: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := migrate.Up(ctx, db.GetDB(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		if a, err := sql.NewPGX(context.Background(), sql.PGXConfig{DB_HOST: host, DB_PORT: port.Port(), DB_USER: "postgres", DB_PASSWORD: "postgres", DB_NAME: "app", DB_MAX_CONNS: 1, DB_MIN_CONNS: 1}); err == nil {
			_, _ = a.GetDB().Exec(context.Background(), "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
			a.Close()
		}
	})
	return db
}

// RabbitMQ คืน AMQP URL ของ broker (container ใช้ร่วมกันทั้ง test binary) — ให้ test ประกาศ queue ของตัวเอง
func RabbitMQ(t testing.TB) string {
	t.Helper()
	RequireDocker(t)
	mqOnce.Do(func() {
		ctx := context.Background()
		c, err := tcrabbit.Run(ctx, RabbitImage, tcrabbit.WithAdminUsername("guest"), tcrabbit.WithAdminPassword("guest"))
		if err != nil {
			mqErr = err
			return
		}
		mqURL, mqErr = c.AmqpURL(ctx)
	})
	if mqErr != nil {
		t.Fatalf("testhelper: start rabbitmq: %v", mqErr)
	}
	return mqURL
}

func randomSuffix() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
