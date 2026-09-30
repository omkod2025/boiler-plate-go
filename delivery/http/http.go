package http

import (
	"context"

	"github.com/omkod2025/boiler-plate-go/configs"
	"github.com/omkod2025/boiler-plate-go/delivery/http/middleware"
	"github.com/omkod2025/boiler-plate-go/delivery/http/routes"

	"github.com/gin-gonic/gin"
	apmgin "go.elastic.co/apm/module/apmgin/v2"
)

// NewRouter สร้าง gin.Engine พร้อม middleware และ routes ทั้งหมดของ HTTP delivery
func NewRouter(ctx context.Context, cfg *configs.Config, jwtConfig middleware.JWTConfig, h routes.Handlers) *gin.Engine {
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
	routes.RegisterHealthRoutes(routerGroup)

	// JWT middleware — route ที่ไม่ต้องใช้ token (public) ให้เพิ่มไว้ที่นี่
	ignorePaths := []middleware.IgnoreRule{
		{Pattern: cfg.Env.APP_PREFIX + "/auth/login", Method: "POST"},
		{Pattern: cfg.Env.APP_PREFIX + "/users", Method: "POST"},
	}
	routerGroup.Use(middleware.JWTOptionalWithIgnoreRules(jwtConfig, ignorePaths))

	// feature routes
	routes.RegisterRoutes(routerGroup, h)

	return r
}
