package handler

import (
	"errors"
	"net/http"

	"github.com/omkod2025-boop/omgon-notification-service/domain/auth"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/response"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/validator"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	useCase *auth.AuthUseCase
}

func NewAuthHandler(useCase *auth.AuthUseCase) *AuthHandler {
	return &AuthHandler{useCase: useCase}
}

// Login เข้าสู่ระบบด้วยอีเมลและรหัสผ่าน (public route)
func (h *AuthHandler) Login(c *gin.Context) {
	var dto LoginRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		if validator.ResponseValidationError(c, err, &dto) {
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	token, err := h.useCase.Login(c.Request.Context(), dto.toInput())
	if errors.Is(err, auth.ErrInvalidCredentials) {
		response.Unauthorized(c, err.Error())
		return
	}
	if err != nil {
		response.Error(c, "failed to login", http.StatusInternalServerError)
		return
	}
	response.Success(c, "ok", toTokenResponse(token))
}
