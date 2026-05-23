package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		start := time.Now()
		c.Next()

		duration := time.Since(start)
		log.Printf("[REQ] id=%s method=%s path=%s status=%d duration=%s ip=%s",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration.Truncate(time.Microsecond),
			c.ClientIP(),
		)
	}
}
