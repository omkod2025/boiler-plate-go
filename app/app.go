package app

import (
	"context"
	"github.com/omkod2025-boop/omgon-notification-service/app/categories"
	"github.com/omkod2025-boop/omgon-notification-service/configs"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/logger"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/middleware"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/response"

	"github.com/gin-gonic/gin"
	apmgin "go.elastic.co/apm/module/apmgin/v2"
)

func InitApp(ctx context.Context, cfg *configs.Config) *gin.Engine {
	// ตรวจสอบ context cancellation
	select {
	case <-ctx.Done():
		logger.Error("Context cancelled during app initialization")
		return nil
	default:
	}

	r := gin.New()

	// middleware
	r.Use(middleware.Logger())
	r.Use(apmgin.Middleware(r))
	r.Use(middleware.APMTracerMiddleware(&ctx))
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(cfg.Env.WHITE_LIST_URL, cfg.Env.ALLOW_HEADERS, cfg.Env.ALLOW_METHODS))
	r.Use(middleware.RateLimitPerMinute(cfg.Env.APP_LIMIT, cfg.Env.APP_LIMIT))

	// router group
	routerGroup := r.Group(cfg.Env.APP_PREFIX)

	// health check
	routerGroup.GET("/healthz", func(c *gin.Context) {
		response.Success(c, "Available", nil)
	})
	routerGroup.GET("/readiness", func(c *gin.Context) {
		response.Success(c, "Ready", nil)
	})

	// โหลด JWT config
	jwtConfig, err := configs.LoadJWTConfig(cfg.Env)
	if err != nil {
		logger.Error("Failed to load JWT config: ", err)
		return nil
	}

	// JWT middleware
	ignorePaths := []middleware.IgnoreRule{
		{Pattern: cfg.Env.APP_PREFIX + "/user-profiles", Method: "POST"},
		{Pattern: cfg.Env.APP_PREFIX + "/user-profiles/.*", Method: "ANY"},
	}
	routerGroup.Use(middleware.JWTOptionalWithIgnoreRules(*jwtConfig, ignorePaths))

	// Register all modules/routes here
	categories.RegisterCategoriesModule(routerGroup, cfg.DB, *jwtConfig)

	logger.Info("Application initialized successfully")
	return r
}

func InitRPCServer(ctx *context.Context, cfg *configs.Config) error {
	// ตรวจสอบ context cancellation
	select {
	case <-(*ctx).Done():
		logger.Info("Context cancelled, skipping RPC server initialization")
		return nil
	default:
	}

	// TODO: Implement RPC server if needed
	logger.Info("RPC server initialized (placeholder)")
	return nil
}
