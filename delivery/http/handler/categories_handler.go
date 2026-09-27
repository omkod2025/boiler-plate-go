package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/omkod2025-boop/omgon-notification-service/delivery/http/middleware"
	"github.com/omkod2025-boop/omgon-notification-service/domain/categories"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/response"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/validator"

	"github.com/gin-gonic/gin"
)

type CategoriesHandler struct {
	useCase *categories.CategoryUseCase
}

func NewCategoriesHandler(useCase *categories.CategoryUseCase) *CategoriesHandler {
	return &CategoriesHandler{useCase: useCase}
}

// currentUserProfileID อ่าน user id จาก JWT และเขียน error response ให้เองเมื่อไม่พบ
func currentUserProfileID(c *gin.Context) (int, bool) {
	userID, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		response.Unauthorized(c, "User not found in context")
		return 0, false
	}
	userProfileID, err := strconv.Atoi(userID)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return 0, false
	}
	return userProfileID, true
}

// pathID อ่าน :id จาก path และเขียน error response ให้เองเมื่อไม่ใช่ตัวเลข
func pathID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id: ต้องเป็นตัวเลข")
		return 0, false
	}
	return id, true
}

// respondUseCaseError แปลง domain error เป็น HTTP status
func respondUseCaseError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, categories.ErrNotFound):
		response.NotFound(c, "category not found")
	case errors.Is(err, categories.ErrInvalidType):
		response.BadRequest(c, err.Error())
	default:
		response.Error(c, fallback, http.StatusInternalServerError)
	}
}

func (h *CategoriesHandler) List(c *gin.Context) {
	userProfileID, ok := currentUserProfileID(c)
	if !ok {
		return
	}

	items, err := h.useCase.List(c.Request.Context(), userProfileID)
	if err != nil {
		respondUseCaseError(c, err, "failed to list categories")
		return
	}
	out := make([]CategoryResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toCategoryResponse(item))
	}
	response.Success(c, "ok", out)
}

func (h *CategoriesHandler) Create(c *gin.Context) {
	userProfileID, ok := currentUserProfileID(c)
	if !ok {
		return
	}

	var dto CreateCategoryRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		if validator.ResponseValidationError(c, err, &dto) {
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.useCase.Create(c.Request.Context(), userProfileID, dto.toInput())
	if err != nil {
		respondUseCaseError(c, err, "failed to create category")
		return
	}
	response.Success(c, "created", toCategoryResponse(item))
}

func (h *CategoriesHandler) Update(c *gin.Context) {
	userProfileID, ok := currentUserProfileID(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	var dto UpdateCategoryRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		if validator.ResponseValidationError(c, err, &dto) {
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.useCase.Update(c.Request.Context(), userProfileID, id, dto.toInput())
	if err != nil {
		respondUseCaseError(c, err, "failed to update category")
		return
	}
	response.Success(c, "updated", toCategoryResponse(item))
}

func (h *CategoriesHandler) Delete(c *gin.Context) {
	userProfileID, ok := currentUserProfileID(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.useCase.Delete(c.Request.Context(), userProfileID, id); err != nil {
		respondUseCaseError(c, err, "failed to delete category")
		return
	}
	response.Success(c, "deleted", nil)
}
