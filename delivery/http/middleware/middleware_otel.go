package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Tracing สร้าง span ต่อ request ด้วย OpenTelemetry (ชื่อ span = route template เช่น "GET /api/categories/:id")
// และต่อ trace จาก header traceparent ของ caller — ตั้งค่า exporter ที่ pkg/telemetry
func Tracing(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}
