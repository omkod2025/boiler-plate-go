package app

import (
	"github.com/omkod2025/boiler-plate-go/delivery/http/handler"
	"github.com/omkod2025/boiler-plate-go/domain/users"
	"github.com/omkod2025/boiler-plate-go/infrastructure/postgres"
	"github.com/omkod2025/boiler-plate-go/infrastructure/security"
)

// newUsersHandler wiring ของ domain users: repository + password hasher -> use case -> handler
func newUsersHandler(d dependencies) *handler.UsersHandler {
	repo := postgres.NewUserRepository(d.db)
	hasher := security.NewBcryptHasher()
	useCase := users.NewUserUseCase(repo, hasher)
	return handler.NewUsersHandler(useCase)
}
