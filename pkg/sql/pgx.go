package sql

import (
	"context"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/logger"
	"fmt"

	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PGX struct {
	pool *pgxpool.Pool
	ctx  context.Context
}

type PGXConfig struct {
	DB_HOST                string
	DB_PORT                string
	DB_USER                string
	DB_PASSWORD            string
	DB_NAME                string
	DB_MAX_CONNS           int
	DB_MIN_CONNS           int
	DB_MAX_CONN_LIFETIME   time.Duration
	DB_MAX_CONN_IDLE_TIME  time.Duration
	DB_HEALTH_CHECK_PERIOD time.Duration
}

func NewPGX(ctx context.Context, config PGXConfig) (*PGX, error) {
	pool, err := Connect(ctx, config)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return nil, err
	}
	return &PGX{
		pool: pool,
		ctx:  ctx,
	}, nil
}

func Connect(ctx context.Context, config PGXConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.DB_USER,
		config.DB_PASSWORD,
		config.DB_HOST,
		config.DB_PORT,
		config.DB_NAME,
	)
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = int32(config.DB_MAX_CONNS)
	poolConfig.MinConns = int32(config.DB_MIN_CONNS)
	poolConfig.MaxConnLifetime = config.DB_MAX_CONN_LIFETIME
	poolConfig.MaxConnIdleTime = config.DB_MAX_CONN_IDLE_TIME
	poolConfig.HealthCheckPeriod = config.DB_HEALTH_CHECK_PERIOD
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func (p *PGX) GetDB() *pgxpool.Pool {
	return p.pool
}

func (p *PGX) Close() {
	p.pool.Close()
}

func (p *PGX) Ping() error {
	return p.pool.Ping(context.Background())
}

// HealthCheck performs a health check on the pool
func (p *PGX) HealthCheck() error {
	return p.pool.Ping(context.Background())
}
