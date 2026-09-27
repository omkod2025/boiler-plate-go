package handler

import (
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

func (h *CategoriesHandler) List(c *gin.Context) {
	userID, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		response.Unauthorized(c, "User not found in context")
		return
	}

	userProfileID, err := strconv.Atoi(userID)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	items, err := h.useCase.List(c.Request.Context(), userProfileID)
	if err != nil {
		response.Error(c, "failed to list categories", http.StatusInternalServerError)
		return
	}
	response.Success(c, "ok", items)
}

func (h *CategoriesHandler) Create(c *gin.Context) {
	userID, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		response.Unauthorized(c, "User not found in context")
		return
	}

	userProfileID, err := strconv.Atoi(userID)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var dto categories.CreateCategoryRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		if validator.ResponseValidationError(c, err, &dto) {
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.useCase.Create(c.Request.Context(), dto, userProfileID)
	if err != nil {
		response.Error(c, "failed to create category", http.StatusInternalServerError)
		return
	}
	response.Success(c, "created", item)
}

func (h *CategoriesHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if err := validator.ValidateVar(id, "required,number"); err != nil {
		response.BadRequest(c, "id: ต้องเป็นตัวเลข")
		return
	}
	var dto categories.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		if validator.ResponseValidationError(c, err, &dto) {
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.useCase.Update(c.Request.Context(), id, dto)
	if err != nil {
		response.Error(c, "failed to update category", http.StatusInternalServerError)
		return
	}
	response.Success(c, "updated", item)
}

func (h *CategoriesHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := validator.ValidateVar(id, "required,number"); err != nil {
		response.BadRequest(c, "id: ต้องเป็นตัวเลข")
		return
	}
	if err := h.useCase.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, "failed to delete category", http.StatusInternalServerError)
		return
	}
	response.Success(c, "deleted", nil)
}
