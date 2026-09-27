package websocket

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register WebSocket endpoints — handler ของ WebSocket ให้วางไว้ใน package นี้
// และเรียกใช้ use case จาก domain/<feature> เช่นเดียวกับ HTTP handler
//
// ตัวอย่าง (ใช้ร่วมกับ library เช่น github.com/gorilla/websocket หรือ github.com/coder/websocket):
//
//	r.GET("/ws", func(c *gin.Context) {
//		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
//		...
//	})
func RegisterRoutes(r *gin.RouterGroup) {
	// TODO: Implement WebSocket handlers if needed
}
