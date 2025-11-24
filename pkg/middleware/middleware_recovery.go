package middleware

import "github.com/gin-gonic/gin"

// Recovery middleware
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}
