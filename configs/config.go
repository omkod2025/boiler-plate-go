package configs

import (
	"context"
	"encoding/base64"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/logger"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/sql"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env EnvConfig
	DB  *sql.PGX
}

type EnvConfig struct {
	APP_PORT              string
	APP_NAME              string
	APP_ENV               string
	APP_VERSION           string
	APP_PREFIX            string
	APP_LIMIT             int
	WHITE_LIST_URL        string
	ALLOW_HEADERS         []string
	ALLOW_METHODS         []string
	LOGGER_LEVEL          string
	JWT_PRIVATE_KEY       string
	JWT_PUBLIC_KEY        string
	JWT_DURATION          time.Duration
	JWT_ISSUER            string
	JWT_AUDIENCE          string
	JWT_VALIDATE_AUDIENCE bool
}

func LoadConfig(ctx context.Context) *Config {
	_ = godotenv.Load()

	dbConfig := sql.PGXConfig{
		DB_HOST:                getEnv("DB_HOST", "localhost", true),
		DB_PORT:                getEnv("DB_PORT", "5432", false),
		DB_USER:                getEnv("DB_USER", "user", true),
		DB_PASSWORD:            getEnv("DB_PASSWORD", "pass", true),
		DB_NAME:                getEnv("DB_NAME", "fmg_auth", true),
		DB_MAX_CONNS:           getEnvInt("DB_MAX_CONNS", 10),
		DB_MIN_CONNS:           getEnvInt("DB_MIN_CONNS", 2),
		DB_MAX_CONN_LIFETIME:   getEnvDuration("DB_MAX_CONN_LIFETIME", 30*time.Minute),
		DB_MAX_CONN_IDLE_TIME:  getEnvDuration("DB_MAX_CONN_IDLE_TIME", 5*time.Minute),
		DB_HEALTH_CHECK_PERIOD: getEnvDuration("DB_HEALTH_CHECK_PERIOD", 1*time.Minute),
	}
	db, err := InitPGXConnection(ctx, dbConfig)
	if err != nil {
		panic(err)
	}

	allowHeaders := []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token", "X-API-KEY", "X-API-SECRET"}
	allowMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	env := EnvConfig{
		APP_PORT:              getEnv("APP_PORT", "21400", true),
		APP_NAME:              getEnv("APP_NAME", "fmg-auth-api", false),
		APP_ENV:               getEnv("APP_ENV", "development", false),
		APP_VERSION:           getEnv("APP_VERSION", "1.0.0", false),
		APP_PREFIX:            getEnv("APP_PREFIX", "/api", false),
		WHITE_LIST_URL:        getEnv("WHITE_LIST_URL", "*", false),
		ALLOW_HEADERS:         allowHeaders,
		ALLOW_METHODS:         allowMethods,
		APP_LIMIT:             getEnvInt("APP_LIMIT", 100),
		LOGGER_LEVEL:          getEnv("LOG_LEVEL", "info", false),
		JWT_PRIVATE_KEY:       getEnvWithBase64Decode("JWT_PRIVATE_KEY", "", true),
		JWT_PUBLIC_KEY:        getEnvWithBase64Decode("JWT_PUBLIC_KEY", "", true),
		JWT_DURATION:          getEnvDuration("JWT_DURATION", 24*time.Hour),
		JWT_ISSUER:            getEnv("JWT_ISSUER", "fmg-auth-api", false),
		JWT_AUDIENCE:          getEnv("JWT_AUDIENCE", "", false),          // ไม่บังคับให้มี audience
		JWT_VALIDATE_AUDIENCE: getEnvBool("JWT_VALIDATE_AUDIENCE", false), // ปิดการตรวจสอบ audience เป็นค่าเริ่มต้น
	}
	level, err := logger.ParseLevel(env.LOGGER_LEVEL)
	if err != nil {
		logger.Warn("invalid LOG_LEVEL, using info: ", err)
	}
	logger.SetLevel(level)

	logger.Info("env", env)
	return &Config{
		Env: env,
		DB:  db,
	}
}

func getEnv(key, fallback string, require bool) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	if require {
		panic("Required environment variable missing: " + key)
	}
	return fallback
}

// getEnvWithBase64Decode ดึง environment variable และ decode base64 ถ้าจำเป็น
func getEnvWithBase64Decode(key, fallback string, require bool) string {
	value := getEnv(key, fallback, require)

	// ตรวจสอบว่าชื่อ key มีคำว่า PRIVATE_KEY หรือ PUBLIC_KEY หรือไม่
	if strings.Contains(strings.ToUpper(key), "PRIVATE_KEY") || strings.Contains(strings.ToUpper(key), "PUBLIC_KEY") {
		// ลอง decode base64
		if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
			return string(decoded)
		}
		// ถ้า decode ไม่ได้ ให้ใช้ค่าเดิม (อาจจะเป็น PEM format ธรรมดา)
		logger.Info("Failed to decode base64 for key: " + key + ", using original value")
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if v, err := strconv.ParseBool(value); err == nil {
			return v
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		d, err := time.ParseDuration(value)
		if err == nil {
			return d
		}
	}
	return fallback
}
