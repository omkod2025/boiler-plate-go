package routes

import (
	"github.com/omkod2025-boop/omgon-notification-service/delivery/http/handler"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handlers รวม handler ทั้งหมดของ HTTP delivery
type Handlers struct {
	Auth       *handler.AuthHandler
	Users      *handler.UsersHandler
	Categories *handler.CategoriesHandler
}

// RegisterHealthRoutes health check สำหรับ liveness/readiness probe
func RegisterHealthRoutes(r *gin.RouterGroup) {
	r.GET("/healthz", func(c *gin.Context) {
		response.Success(c, "Available", nil)
	})
	r.GET("/readiness", func(c *gin.Context) {
		response.Success(c, "Ready", nil)
	})
}

// RegisterRoutes register routes ของทุก feature ที่นี่
func RegisterRoutes(r *gin.RouterGroup, h Handlers) {
	RegisterAuthRoutes(r, h.Auth)
	RegisterUsersRoutes(r, h.Users)
	RegisterCategoriesRoutes(r, h.Categories)
}
