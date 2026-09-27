package routes

import (
	"github.com/omkod2025-boop/omgon-notification-service/delivery/http/handler"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.RouterGroup, h *handler.AuthHandler) {
	group := r.Group("/auth")
	{
		group.POST("/login", h.Login)
	}
}
