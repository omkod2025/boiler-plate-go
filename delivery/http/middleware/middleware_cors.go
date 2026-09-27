package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS middleware (configurable)
func CORS(whiteListURL string, allowHeaders []string, allowMethods []string) gin.HandlerFunc {
	origins := []string{"*"}
	if whiteListURL != "" {
		origins = strings.Split(whiteListURL, ",")
	}
	corsConfig := cors.Config{
		AllowOrigins:     origins,
		AllowHeaders:     allowHeaders,
		AllowMethods:     allowMethods,
		AllowCredentials: true,
	}
	return cors.New(corsConfig)
}
