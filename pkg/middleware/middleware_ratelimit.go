package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*clientLimiter)
)

// RateLimitPerMinute สร้าง middleware limit ต่อ 1 นาที ต่อ IP
func RateLimitPerMinute(limitPerMin, burst int) gin.HandlerFunc {
	cleanupInterval := time.Minute
	if burst <= 0 {
		burst = limitPerMin
	}
	go func() {
		for {
			time.Sleep(cleanupInterval)
			mu.Lock()
			for ip, cl := range clients {
				if time.Since(cl.lastSeen) > 5*cleanupInterval {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		cl, exists := clients[ip]
		if !exists {
			limiter := rate.NewLimiter(rate.Limit(float64(limitPerMin)/60), burst)
			cl = &clientLimiter{limiter: limiter, lastSeen: time.Now()}
			clients[ip] = cl
		}
		cl.lastSeen = time.Now()
		mu.Unlock()
		if !cl.limiter.Allow() {
			c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
