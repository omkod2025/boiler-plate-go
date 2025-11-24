package categories

import "github.com/gin-gonic/gin"

func RegisterCategoriesRoutes(r *gin.RouterGroup, handler *CategoriesHandler) {
	group := r.Group("/categories")
	{
		group.GET("", handler.List)
		group.POST("", handler.Create)
		group.PUT(":id", handler.Update)
		group.DELETE(":id", handler.Delete)
	}
}
