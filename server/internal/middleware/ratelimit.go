package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func RateLimit(maxPerMin int) gin.HandlerFunc {
	mu := sync.Mutex{}
	hits := make(map[string][]time.Time)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		now := time.Now()
		cutoff := now.Add(-time.Minute)
		filtered := hits[ip][:0]
		for _, t := range hits[ip] {
			if t.After(cutoff) {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) >= maxPerMin {
			mu.Unlock()
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "RATE_LIMITED", "message": "too many requests"},
			})
			return
		}
		hits[ip] = append(filtered, now)
		mu.Unlock()
		c.Next()
	}
}
