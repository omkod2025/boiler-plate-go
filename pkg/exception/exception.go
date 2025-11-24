package exception

import "net/http"

// AppError สำหรับ custom error ที่กำหนด code, message, httpStatus ได้
// Example: exception.NewAppError(500, "connect db failed")
type AppError struct {
	Code      int    // HTTP status code เช่น 404, 500
	Message   string // error message
	ErrorCode string // (optional) business error code เช่น "DB_CONN_FAIL"
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError สร้าง custom error
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewAppErrorWithCode สร้าง custom error พร้อม business error code
func NewAppErrorWithCode(code int, message, errorCode string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		ErrorCode: errorCode,
	}
}

// ตัวอย่าง error helper
func InternalError() *AppError {
	message := "Internal Server Error"
	return NewAppError(http.StatusInternalServerError, message)
}

func NotFoundError(message string) *AppError {
	return NewAppError(http.StatusNotFound, message)
}

func BadRequestError(message string) *AppError {
	return NewAppError(http.StatusBadRequest, message)
}

func ForbiddenError(message string) *AppError {
	return NewAppError(http.StatusForbidden, message)
}
