package app

import (
	"github.com/omkod2025-boop/omgon-notification-service/delivery/http/handler"
	"github.com/omkod2025-boop/omgon-notification-service/domain/auth"
	"github.com/omkod2025-boop/omgon-notification-service/infrastructure/postgres"
	"github.com/omkod2025-boop/omgon-notification-service/infrastructure/security"
)

// newAuthHandler wiring ของ domain auth: credential repository + password verifier + token issuer -> use case -> handler
func newAuthHandler(d dependencies) *handler.AuthHandler {
	credentials := postgres.NewCredentialRepository(d.db)
	verifier := security.NewBcryptHasher()
	tokens := security.NewJWTTokenIssuer(d.jwt)
	useCase := auth.NewAuthUseCase(credentials, verifier, tokens)
	return handler.NewAuthHandler(useCase)
}
