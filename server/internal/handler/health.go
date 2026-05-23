package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查接口处理器
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler 创建健康检查处理器实例
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check godoc
// @Summary      健康检查
// @Tags         health
// @Success      200  {object}  object{status=string}
// @Failure      503  {object}  object{status=string,error=string}
// @Router       /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	// 通过 Ping 数据库来判断服务是否健康
	if err := h.db.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
