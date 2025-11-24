package gin

import (
	"github.com/gin-gonic/gin"
)

// NewEngine สร้าง gin.Engine พร้อม custom middleware หรือ config ที่ต้องการ
func NewEngine() *gin.Engine {
	r := gin.New()

	// ตัวอย่าง: เพิ่ม logger, recovery middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// TODO: เพิ่ม custom middleware, error handler, หรือ config อื่น ๆ ที่นี่

	return r
}

// สามารถเพิ่ม utility function อื่น ๆ สำหรับ gin ได้ที่นี่
