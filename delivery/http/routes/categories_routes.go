package routes

import (
	"github.com/omkod2025-boop/omgon-notification-service/delivery/http/handler"

	"github.com/gin-gonic/gin"
)

func RegisterCategoriesRoutes(r *gin.RouterGroup, h *handler.CategoriesHandler) {
	group := r.Group("/categories")
	{
		group.GET("", h.List)
		group.POST("", h.Create)
		group.PUT(":id", h.Update)
		group.DELETE(":id", h.Delete)
	}
}
