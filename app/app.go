package app

import (
	"context"

	"github.com/omkod2025-boop/omgon-notification-service/configs"
	httpdelivery "github.com/omkod2025-boop/omgon-notification-service/delivery/http"
	"github.com/omkod2025-boop/omgon-notification-service/delivery/http/handler"
	"github.com/omkod2025-boop/omgon-notification-service/delivery/http/middleware"
	"github.com/omkod2025-boop/omgon-notification-service/delivery/http/routes"
	"github.com/omkod2025-boop/omgon-notification-service/delivery/rpc"
	"github.com/omkod2025-boop/omgon-notification-service/domain/categories"
	"github.com/omkod2025-boop/omgon-notification-service/infrastructure/postgres"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/jwt"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

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

	// infrastructure -> domain use case
	categoryRepo := postgres.NewCategoryRepository(cfg.DB)
	categoryUseCase := categories.NewCategoryUseCase(categoryRepo)

	// delivery: handler
	handlers := routes.Handlers{
		Categories: handler.NewCategoriesHandler(categoryUseCase),
	}

	r := httpdelivery.NewRouter(ctx, cfg, jwtConfig, handlers)

	logger.Info("Application initialized successfully")
	return r
}

func InitRPCServer(ctx *context.Context, cfg *configs.Config) error {
	return rpc.InitRPCServer(*ctx, cfg)
}

// loadJWTConfig โหลด JWT configuration พร้อม RSA keys
func loadJWTConfig(env configs.EnvConfig) (middleware.JWTConfig, error) {
	privateKey, err := jwt.LoadRSAPrivateKey(env.JWT_PRIVATE_KEY)
	if err != nil {
		return middleware.JWTConfig{}, err
	}

	publicKey, err := jwt.LoadRSAPublicKey(env.JWT_PUBLIC_KEY)
	if err != nil {
		return middleware.JWTConfig{}, err
	}

	return middleware.JWTConfig{
		PrivateKey:       privateKey,
		PublicKey:        publicKey,
		TokenDuration:    env.JWT_DURATION,
		Issuer:           env.JWT_ISSUER,
		Audience:         env.JWT_AUDIENCE,
		ValidateAudience: env.JWT_VALIDATE_AUDIENCE,
	}, nil
}
