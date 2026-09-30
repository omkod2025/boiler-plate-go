package configs

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/omkod2025/boiler-plate-go/pkg/logger"
	"github.com/omkod2025/boiler-plate-go/pkg/sql"

	"github.com/joho/godotenv"
)

// Config ของ process หนึ่งตัว — DB เป็น nil สำหรับ role ที่ไม่ใช้ database (receiver, sandbox)
type Config struct {
	Env EnvConfig
	DB  *sql.PGX
}

type EnvConfig struct {
	APP_ROLE              string
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
	SHUTDOWN_TIMEOUT      time.Duration
	JWT_PRIVATE_KEY       string
	JWT_PUBLIC_KEY        string
	JWT_DURATION          time.Duration
	JWT_ISSUER            string
	JWT_AUDIENCE          string
	JWT_VALIDATE_AUDIENCE bool
	AMQP_URL              string
	AMQP_QUEUE            string
	AMQP_EXCHANGE         string
	AMQP_PREFETCH         int
	OTEL_ENDPOINT         string
}

// String ใช้เมื่อ log หรือ print config — ปิดค่า secret ไว้เสมอ เพื่อไม่ให้ key หลุดไปใน log
func (e EnvConfig) String() string {
	// แปลงเป็น type ที่ไม่มี method String เพื่อไม่ให้ fmt เรียก String ซ้ำจนวนไม่รู้จบ
	type plain EnvConfig
	redacted := plain(e)
	redacted.JWT_PRIVATE_KEY = redact(e.JWT_PRIVATE_KEY)
	redacted.JWT_PUBLIC_KEY = redact(e.JWT_PUBLIC_KEY)
	redacted.AMQP_URL = redact(e.AMQP_URL) // มีรหัสผ่านของ broker
	return fmt.Sprintf("%+v", redacted)
}

// GoString ใช้กับ %#v — ปิดค่า secret เหมือน String
func (e EnvConfig) GoString() string {
	return e.String()
}

func redact(value string) string {
	if value == "" {
		return ""
	}
	return "[REDACTED]"
}

// LoadConfig อ่าน environment variable ทั้งหมด (และ .env ถ้ามี) แต่ยังไม่ต่อ database
// role ที่ใช้ DB เรียก ConnectDB ต่อเอง (ดู app/roles.go)
func LoadConfig(role string) *Config {
	_ = godotenv.Load()

	allowHeaders := []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token", "X-API-KEY", "X-API-SECRET"}
	allowMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	env := EnvConfig{
		APP_ROLE:              role,
		APP_PORT:              getEnv("APP_PORT", "21400", false),
		APP_NAME:              getEnv("APP_NAME", "boiler-plate-go", false),
		APP_ENV:               getEnv("APP_ENV", "development", false),
		APP_VERSION:           getEnv("APP_VERSION", "1.0.0", false),
		APP_PREFIX:            getEnv("APP_PREFIX", "/api", false),
		WHITE_LIST_URL:        getEnv("WHITE_LIST_URL", "*", false),
		ALLOW_HEADERS:         allowHeaders,
		ALLOW_METHODS:         allowMethods,
		APP_LIMIT:             getEnvInt("APP_LIMIT", 100),
		LOGGER_LEVEL:          getEnv("LOG_LEVEL", "info", false),
		SHUTDOWN_TIMEOUT:      getEnvDuration("SHUTDOWN_TIMEOUT", 20*time.Second),
		JWT_PRIVATE_KEY:       getEnvWithBase64Decode("JWT_PRIVATE_KEY", "", false),
		JWT_PUBLIC_KEY:        getEnvWithBase64Decode("JWT_PUBLIC_KEY", "", false),
		JWT_DURATION:          getEnvDuration("JWT_DURATION", 24*time.Hour),
		JWT_ISSUER:            getEnv("JWT_ISSUER", "boiler-plate-go", false),
		JWT_AUDIENCE:          getEnv("JWT_AUDIENCE", "", false),          // ไม่บังคับให้มี audience
		JWT_VALIDATE_AUDIENCE: getEnvBool("JWT_VALIDATE_AUDIENCE", false), // ปิดการตรวจสอบ audience เป็นค่าเริ่มต้น
		AMQP_URL:              getEnv("AMQP_URL", "", false),
		AMQP_QUEUE:            getEnv("AMQP_QUEUE", "", false),
		AMQP_EXCHANGE:         getEnv("AMQP_EXCHANGE", "", false),
		AMQP_PREFETCH:         getEnvInt("AMQP_PREFETCH", 10),
		OTEL_ENDPOINT:         getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "", false),
	}
	level, err := logger.ParseLevel(env.LOGGER_LEVEL)
	if err != nil {
		logger.Warn("invalid LOG_LEVEL, using info: ", err)
	}
	logger.SetLevel(level)

	logger.Info("env", env)
	return &Config{Env: env}
}

// RequireJWT ตรวจว่ามี key ครบสำหรับ role ที่ออก/ตรวจ token (api, stream)
func (c *Config) RequireJWT() error {
	if c.Env.JWT_PRIVATE_KEY == "" || c.Env.JWT_PUBLIC_KEY == "" {
		return errors.New("JWT_PRIVATE_KEY and JWT_PUBLIC_KEY are required for role " + c.Env.APP_ROLE)
	}
	return nil
}

// ConnectDB อ่านค่า DB_* แล้วเปิด connection pool ใส่ c.DB (role ที่ใช้ database เท่านั้น)
func (c *Config) ConnectDB(ctx context.Context) error {
	dbConfig := sql.PGXConfig{
		DB_HOST:                getEnv("DB_HOST", "", false),
		DB_PORT:                getEnv("DB_PORT", "5432", false),
		DB_USER:                getEnv("DB_USER", "", false),
		DB_PASSWORD:            getEnv("DB_PASSWORD", "", false),
		DB_NAME:                getEnv("DB_NAME", "", false),
		DB_MAX_CONNS:           getEnvInt("DB_MAX_CONNS", 10),
		DB_MIN_CONNS:           getEnvInt("DB_MIN_CONNS", 2),
		DB_MAX_CONN_LIFETIME:   getEnvDuration("DB_MAX_CONN_LIFETIME", 30*time.Minute),
		DB_MAX_CONN_IDLE_TIME:  getEnvDuration("DB_MAX_CONN_IDLE_TIME", 5*time.Minute),
		DB_HEALTH_CHECK_PERIOD: getEnvDuration("DB_HEALTH_CHECK_PERIOD", 1*time.Minute),
	}
	var missing []string
	for k, v := range map[string]string{"DB_HOST": dbConfig.DB_HOST, "DB_USER": dbConfig.DB_USER, "DB_PASSWORD": dbConfig.DB_PASSWORD, "DB_NAME": dbConfig.DB_NAME} {
		if v == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("role %s needs a database; missing %s", c.Env.APP_ROLE, strings.Join(missing, ", "))
	}
	db, err := InitPGXConnection(ctx, dbConfig)
	if err != nil {
		return err
	}
	c.DB = db
	return nil
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
	if value == "" {
		return value
	}

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
