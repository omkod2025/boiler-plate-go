package middleware

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger middleware: log IP address ของ client และ latency เป็น ms, timestamp มี millisecond
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latencyMs := float64(time.Since(start).Nanoseconds()) / 1e6
		timestamp := time.Now().Format(time.RFC3339Nano)
		logLine := fmt.Sprintf("%s | %3d | %9.3fms | %15s | %-7s %s | %s\n",
			timestamp,
			c.Writer.Status(),
			latencyMs,
			c.ClientIP(),
			c.Request.Method,
			c.Request.URL.Path,
			c.Errors.ByType(gin.ErrorTypePrivate).String(),
		)
		os.Stdout.WriteString(logLine)
	}
}