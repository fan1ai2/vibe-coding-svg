package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimit(redisAddr string, maxPerMin int) gin.HandlerFunc {
	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx := context.Background()

	// 启动时验证 Redis 连通性
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("[WARN] rate limiter: redis unreachable at %s, rate limiting disabled: %v", redisAddr, err)
		// 返回一个不做限流的空操作中间件
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		window := time.Now().Unix() / 60
		key := fmt.Sprintf("ratelimit:%s:%d", ip, window)

		count, err := client.Incr(ctx, key).Result()
		if err != nil {
			log.Printf("[ERROR] rate limiter incr: %v", err)
			c.Next()
			return
		}
		if count == 1 {
			client.Expire(ctx, key, 60*time.Second)
		}
		if count > int64(maxPerMin) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "RATE_LIMITED", "message": "请求过于频繁，请稍后重试"},
			})
			return
		}
		c.Next()
	}
}
