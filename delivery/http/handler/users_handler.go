package handler

import (
	"errors"
	"net/http"

	"github.com/omkod2025/boiler-plate-go/domain/users"
	"github.com/omkod2025/boiler-plate-go/pkg/response"
	"github.com/omkod2025/boiler-plate-go/pkg/validator"

	"github.com/gin-gonic/gin"
)

type UsersHandler struct {
	useCase *users.UserUseCase
}

func NewUsersHandler(useCase *users.UserUseCase) *UsersHandler {
	return &UsersHandler{useCase: useCase}
}

// respondUserError แปลง domain error เป็น HTTP status
func respondUserError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, users.ErrNotFound):
		response.NotFound(c, "user not found")
	case errors.Is(err, users.ErrEmailAlreadyExists):
		response.Error(c, err.Error(), http.StatusConflict)
	default:
		response.Error(c, fallback, http.StatusInternalServerError)
	}
}

// Register สมัครสมาชิก (public route)
func (h *UsersHandler) Register(c *gin.Context) {
	var dto RegisterUserRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		if validator.ResponseValidationError(c, err, &dto) {
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	user, err := h.useCase.Register(c.Request.Context(), dto.toInput())
	if err != nil {
		respondUserError(c, err, "failed to register user")
		return
	}
	response.Success(c, "created", toUserResponse(user))
}

// Me ดึงข้อมูลของผู้ใช้ที่ login อยู่
func (h *UsersHandler) Me(c *gin.Context) {
	userProfileID, ok := currentUserProfileID(c)
	if !ok {
		return
	}
	user, err := h.useCase.GetProfile(c.Request.Context(), userProfileID)
	if err != nil {
		respondUserError(c, err, "failed to get user")
		return
	}
	response.Success(c, "ok", toUserResponse(user))
}
