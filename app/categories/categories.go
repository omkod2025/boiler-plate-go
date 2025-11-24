package categories

import (
	"github.com/omkod2025-boop/omgon-notification-service/pkg/middleware"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/sql"

	"github.com/gin-gonic/gin"
)

func RegisterCategoriesModule(r *gin.RouterGroup, db *sql.PGX, _ middleware.JWTConfig) {
	repo := NewCategoryRepository(db)
	useCase := NewCategoryUseCase(repo)
	handler := NewCategoriesHandler(useCase)
	RegisterCategoriesRoutes(r, handler)
}
