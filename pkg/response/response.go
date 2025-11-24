package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	statusSuccess      = "success"
	statusError        = "error"
	statusBadRequest   = "bad_request"
	statusUnauthorized = "unauthorized"
	statusNotFound     = "not_found"
)

type Response struct {
	Success    bool        `json:"success"`
	Status     string      `json:"status"`
	StatusCode int         `json:"statusCode"`
	Message    string      `json:"message"`
	Result     interface{} `json:"result,omitempty"`
}

func Success(c *gin.Context, message string, result interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:    true,
		Status:     statusSuccess,
		StatusCode: http.StatusOK,
		Message:    message,
		Result:     result,
	})
}

func Error(c *gin.Context, message string, statusCodes ...int) {
	statusCode := http.StatusInternalServerError
	if len(statusCodes) > 0 {
		statusCode = statusCodes[0]
	}
	c.JSON(statusCode, Response{
		Success:    false,
		Status:     statusError,
		StatusCode: statusCode,
		Message:    message,
	})
}

func BadRequest(c *gin.Context, message string, statusCodes ...int) {
	statusCode := http.StatusBadRequest
	if len(statusCodes) > 0 {
		statusCode = statusCodes[0]
	}
	c.JSON(http.StatusBadRequest, Response{
		Success:    false,
		Status:     statusBadRequest,
		StatusCode: statusCode,
		Message:    message,
	})
}

func BadRequestWithData(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success":    false,
		"status":     statusBadRequest,
		"statusCode": http.StatusBadRequest,
		"message":    message,
		"data":       data,
	})
}

func Unauthorized(c *gin.Context, message string, statusCodes ...int) {
	statusCode := http.StatusUnauthorized
	if len(statusCodes) > 0 {
		statusCode = statusCodes[0]
	}
	c.JSON(http.StatusUnauthorized, Response{
		Success:    false,
		Status:     statusUnauthorized,
		StatusCode: statusCode,
		Message:    message,
	})
}

func NotFound(c *gin.Context, message string, statusCodes ...int) {
	statusCode := http.StatusNotFound
	if len(statusCodes) > 0 {
		statusCode = statusCodes[0]
	}
	c.JSON(http.StatusNotFound, Response{
		Success:    false,
		Status:     statusNotFound,
		StatusCode: statusCode,
		Message:    message,
	})
}
