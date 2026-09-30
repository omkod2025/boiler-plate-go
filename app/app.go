package app

import (
	"context"

	"github.com/omkod2025/boiler-plate-go/configs"
	httpdelivery "github.com/omkod2025/boiler-plate-go/delivery/http"
	"github.com/omkod2025/boiler-plate-go/delivery/http/routes"
	"github.com/omkod2025/boiler-plate-go/delivery/rpc"
	"github.com/omkod2025/boiler-plate-go/pkg/jwt"
	"github.com/omkod2025/boiler-plate-go/pkg/logger"
	"github.com/omkod2025/boiler-plate-go/pkg/sql"

	"github.com/gin-gonic/gin"
)

// dependencies ของที่ใช้ร่วมกันทุก domain — สร้างครั้งเดียวใน InitApp แล้วส่งให้ wiring ของแต่ละ domain
type dependencies struct {
	cfg *configs.Config
	db  *sql.PGX
	jwt jwt.Config
}

// InitApp ประกอบ application ทั้งหมด
//
// การเพิ่ม domain ใหม่:
//  1. สร้างไฟล์ app/<domain>.go ที่มีฟังก์ชัน new<Domain>Handler(d dependencies)
//     สำหรับ new repository -> use case -> handler ของ domain นั้น
//  2. เพิ่ม field ใน routes.Handlers และเรียก new<Domain>Handler ด้านล่าง
func InitApp(ctx context.Context, cfg *configs.Config) *gin.Engine {
	// ตรวจสอบ context cancellation
	select {
	case <-ctx.Done():
		logger.Error("Context cancelled during app initialization")
		return nil
	default:
	}

	// โหลด JWT config
	jwtConfig, err := loadJWTConfig(cfg.Env)
	if err != nil {
		logger.Error("Failed to load JWT config: ", err)
		return nil
	}

	d := dependencies{cfg: cfg, db: cfg.DB, jwt: jwtConfig}

	handlers := routes.Handlers{
		Auth:       newAuthHandler(d),
		Users:      newUsersHandler(d),
		Categories: newCategoriesHandler(d),
	}

	r := httpdelivery.NewRouter(ctx, cfg, jwtConfig, handlers)

	logger.Info("Application initialized successfully")
	return r
}

func InitRPCServer(ctx *context.Context, cfg *configs.Config) error {
	return rpc.InitRPCServer(*ctx, cfg)
}

// loadJWTConfig โหลด JWT configuration พร้อม RSA keys
func loadJWTConfig(env configs.EnvConfig) (jwt.Config, error) {
	privateKey, err := jwt.LoadRSAPrivateKey(env.JWT_PRIVATE_KEY)
	if err != nil {
		return jwt.Config{}, err
	}

	publicKey, err := jwt.LoadRSAPublicKey(env.JWT_PUBLIC_KEY)
	if err != nil {
		return jwt.Config{}, err
	}

	return jwt.Config{
		PrivateKey:       privateKey,
		PublicKey:        publicKey,
		TokenDuration:    env.JWT_DURATION,
		Issuer:           env.JWT_ISSUER,
		Audience:         env.JWT_AUDIENCE,
		ValidateAudience: env.JWT_VALIDATE_AUDIENCE,
	}, nil
}
