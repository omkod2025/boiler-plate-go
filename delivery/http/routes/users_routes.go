package routes

import (
	"github.com/omkod2025/boiler-plate-go/delivery/http/handler"

	"github.com/gin-gonic/gin"
)

func RegisterUsersRoutes(r *gin.RouterGroup, h *handler.UsersHandler) {
	group := r.Group("/users")
	{
		group.POST("", h.Register)
		group.GET("/me", h.Me)
	}
}
