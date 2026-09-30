package app

import (
	"github.com/omkod2025/boiler-plate-go/delivery/http/handler"
	"github.com/omkod2025/boiler-plate-go/domain/auth"
	"github.com/omkod2025/boiler-plate-go/infrastructure/postgres"
	"github.com/omkod2025/boiler-plate-go/infrastructure/security"
)

// newAuthHandler wiring ของ domain auth: credential repository + password verifier + token issuer -> use case -> handler
func newAuthHandler(d dependencies) *handler.AuthHandler {
	credentials := postgres.NewCredentialRepository(d.db)
	verifier := security.NewBcryptHasher()
	tokens := security.NewJWTTokenIssuer(d.jwt)
	useCase := auth.NewAuthUseCase(credentials, verifier, tokens)
	return handler.NewAuthHandler(useCase)
}
