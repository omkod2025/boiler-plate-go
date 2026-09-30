// Package health ให้ endpoint สำหรับ orchestrator ทุก role:
//
//   - GET /health/live   process ยังทำงาน (ไม่แตะ DB หรือระบบภายนอก)
//   - GET /health/ready  dependency ที่จำเป็นตอบได้ ไม่งั้นตอบ 503 — ตอบแค่ชื่อ check กับสถานะ ไม่มีรายละเอียด error
package health

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Check คือการตรวจ dependency หนึ่งตัว Essential=false จะรายงาน "degraded" แต่ไม่ทำให้ ready ล้ม
type Check struct {
	Name      string
	Essential bool
	Probe     func(ctx context.Context) error
}

// Register ผูก /health/live และ /health/ready
func Register(r gin.IRoutes, checks ...Check) {
	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/health/ready", func(c *gin.Context) {
		status, results := Evaluate(c.Request.Context(), checks)
		code := http.StatusOK
		if status != "ready" {
			code = http.StatusServiceUnavailable
		}
		c.JSON(code, gin.H{"status": status, "checks": results})
	})
}

// Evaluate รันทุก check พร้อมกัน (แต่ละตัวไม่เกิน 2 วินาที)
func Evaluate(ctx context.Context, checks []Check) (string, map[string]string) {
	status := "ready"
	results := make(map[string]string, len(checks))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, chk := range checks {
		wg.Go(func() {
			cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			err := chk.Probe(cctx)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				results[chk.Name] = "ok"
			case chk.Essential:
				results[chk.Name] = "fail"
				status = "unavailable"
			default:
				results[chk.Name] = "degraded"
			}
		})
	}
	wg.Wait()
	return status, results
}
