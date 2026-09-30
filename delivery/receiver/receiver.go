// Package receiver คือ delivery ของ role receiver: รับ webhook ขาเข้า แล้ว publish raw body เข้า
// RabbitMQ (รอ publisher confirm) ก่อนตอบ 2xx — ไม่มี DB และไม่มี business logic
//
// ตรวจ signature ของ provider ให้ใส่ใน verify ของแต่ละ source ก่อน publish (ตัวอย่างนี้ยังไม่ตรวจ)
package receiver

import (
	"context"
	"io"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/omkod2025/boiler-plate-go/pkg/logger"
	"github.com/omkod2025/boiler-plate-go/pkg/response"
)

// Publisher คือส่วนที่ receiver ต้องใช้จาก pkg/amqp
type Publisher interface {
	Publish(ctx context.Context, exchange, routingKey, messageID string, body []byte) error
}

// MaxBody คือขนาด body สูงสุดที่รับ
const MaxBody = 1 << 20

var sourcePattern = regexp.MustCompile(`^[a-z0-9-]{1,40}$`)

// RegisterRoutes ผูก POST /hooks/:source — publish ด้วย routing key hooks.<source>
// pub เป็น nil ได้ (dev ที่ไม่มี RabbitMQ): รับแล้ว log อย่างเดียว
func RegisterRoutes(r *gin.RouterGroup, pub Publisher, exchange string) {
	r.POST("/hooks/:source", func(c *gin.Context) {
		source := c.Param("source")
		if !sourcePattern.MatchString(source) {
			response.NotFound(c, "unknown source")
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, MaxBody))
		if err != nil {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"success": false, "message": "payload too large"})
			return
		}
		id := uuid.Must(uuid.NewV7()).String()
		if pub == nil {
			logger.Info("receiver: accepted " + source + " webhook " + id + " (no broker configured)")
			c.JSON(http.StatusAccepted, gin.H{"receipt_id": id})
			return
		}
		if err := pub.Publish(c.Request.Context(), exchange, "hooks."+source, id, body); err != nil {
			// provider จะส่งซ้ำเมื่อได้ 5xx จึงไม่มีข้อความหาย
			logger.Error("receiver: publish failed: " + err.Error())
			response.Error(c, "temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"receipt_id": id})
	})
}
