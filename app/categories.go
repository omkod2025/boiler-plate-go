package app

import (
	"github.com/omkod2025-boop/omgon-notification-service/delivery/http/handler"
	"github.com/omkod2025-boop/omgon-notification-service/domain/categories"
	"github.com/omkod2025-boop/omgon-notification-service/infrastructure/postgres"
)

// newCategoriesHandler wiring ของ domain categories: repository -> use case -> handler
func newCategoriesHandler(d dependencies) *handler.CategoriesHandler {
	repo := postgres.NewCategoryRepository(d.db)
	useCase := categories.NewCategoryUseCase(repo)
	return handler.NewCategoriesHandler(useCase)
}
