package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db          *sql.DB
	redisClient *redis.Client
	minioClient *minio.Client
}

func NewHealthHandler(db *sql.DB, redisAddr string, minioClient *minio.Client) *HealthHandler {
	rc := redis.NewClient(&redis.Options{Addr: redisAddr})
	return &HealthHandler{db: db, redisClient: rc, minioClient: minioClient}
}

// Check godoc
// @Summary      健康检查
// @Tags         health
// @Success      200  {object}  object{status=string,checks=object}
// @Failure      503  {object}  object{status=string,checks=object}
// @Router       /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}

	// PostgreSQL
	if err := h.db.PingContext(ctx); err != nil {
		checks["postgres"] = "down: " + err.Error()
	} else {
		checks["postgres"] = "ok"
	}

	// Redis
	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		checks["redis"] = "down: " + err.Error()
	} else {
		checks["redis"] = "ok"
	}

	// MinIO
	_, err := h.minioClient.ListBuckets(ctx)
	if err != nil {
		checks["minio"] = "down: " + err.Error()
	} else {
		checks["minio"] = "ok"
	}

	// 判断整体健康状态
	healthy := true
	for _, v := range checks {
		if v != "ok" {
			healthy = false
			break
		}
	}

	if healthy {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "checks": checks})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "checks": checks})
	}
}
