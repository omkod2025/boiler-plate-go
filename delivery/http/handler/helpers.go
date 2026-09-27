package handler

import (
	"strconv"

	"github.com/omkod2025-boop/omgon-notification-service/delivery/http/middleware"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/response"

	"github.com/gin-gonic/gin"
)

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
