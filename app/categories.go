package app

import (
	"github.com/omkod2025/boiler-plate-go/delivery/http/handler"
	"github.com/omkod2025/boiler-plate-go/domain/categories"
	"github.com/omkod2025/boiler-plate-go/infrastructure/postgres"
)

// newCategoriesHandler wiring ของ domain categories: repository -> use case -> handler
func newCategoriesHandler(d dependencies) *handler.CategoriesHandler {
	repo := postgres.NewCategoryRepository(d.db)
	useCase := categories.NewCategoryUseCase(repo)
	return handler.NewCategoriesHandler(useCase)
}
