package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	repo *repo.TagRepo
}

func NewTagHandler(r *repo.TagRepo) *TagHandler {
	return &TagHandler{repo: r}
}

func (h *TagHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	sort := c.DefaultQuery("sort", "popular")
	tags, err := h.repo.List(sort, limit)
	if err != nil {
		log.Printf("[ERROR] list tags: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "LIST_FAILED", "message": "获取标签列表失败"}})
		return
	}
	if tags == nil {
		tags = make([]model.Tag, 0)
	}
	c.JSON(http.StatusOK, gin.H{"data": tags})
}
