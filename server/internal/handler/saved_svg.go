package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/gin-gonic/gin"
)

type SavedSvgHandler struct {
	repo *repo.SavedSvgRepo
}

func NewSavedSvgHandler(r *repo.SavedSvgRepo) *SavedSvgHandler {
	return &SavedSvgHandler{repo: r}
}

type saveSvgRequest struct {
	Name       string `json:"name" binding:"required"`
	SvgContent string `json:"svg_content" binding:"required"`
}

func (h *SavedSvgHandler) Save(c *gin.Context) {
	userID := c.GetString("user_id")
	var req saveSvgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_BODY", "message": "请提供 name 和 svg_content"}})
		return
	}
	if len(req.SvgContent) > 2*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"code": "TOO_LARGE", "message": "SVG 内容超过 2MB 限制"}})
		return
	}
	s, err := h.repo.Create(userID, req.Name, req.SvgContent)
	if err != nil {
		log.Printf("[ERROR] save svg user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "SAVE_FAILED", "message": "保存失败"}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": s})
}

func (h *SavedSvgHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	list, err := h.repo.FindByUserID(userID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] list svgs user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "LIST_FAILED", "message": "获取列表失败"}})
		return
	}
	if list == nil {
		list = make([]*model.SavedSvg, 0)
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *SavedSvgHandler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	s, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": "查询失败"}})
		return
	}
	if s == nil || s.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "记录不存在"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": s})
}

func (h *SavedSvgHandler) Download(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	s, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": "查询失败"}})
		return
	}
	if s == nil || s.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "记录不存在"}})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+s.Name+".svg\"")
	c.Header("Content-Type", "image/svg+xml")
	c.String(http.StatusOK, s.SvgContent)
}

func (h *SavedSvgHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	// 验证记录属于当前用户
	s, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": "查询失败"}})
		return
	}
	if s == nil || s.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "记录不存在"}})
		return
	}
	if err := h.repo.Delete(id); err != nil {
		log.Printf("[ERROR] delete svg id=%s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DELETE_FAILED", "message": "删除失败"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true}})
}
