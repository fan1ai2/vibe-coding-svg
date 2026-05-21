package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)

type ConversionHandler struct {
	cfg *config.Config
	svc *service.ConversionService
}

func NewConversionHandler(cfg *config.Config, svc *service.ConversionService) *ConversionHandler {
	return &ConversionHandler{cfg: cfg, svc: svc}
}

func (h *ConversionHandler) Upload(c *gin.Context) {
	userID := c.GetString("user_id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "NO_FILE", "message": "file is required"}})
		return
	}
	defer file.Close()

	if header.Size > h.cfg.MaxFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"code": "FILE_TOO_LARGE", "message": "file exceeds maximum size"}})
		return
	}

	conv, err := h.svc.Enqueue(userID, file, header.Filename, header.Size)
	if err != nil {
		if strings.Contains(err.Error(), "quota") {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "QUOTA_EXCEEDED", "message": err.Error()}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "UPLOAD_FAILED", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": conv})
}

func (h *ConversionHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	list, err := h.svc.List(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "LIST_FAILED", "message": err.Error()}})
		return
	}
	if list == nil {
		list = make([]*model.Conversion, 0)
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *ConversionHandler) Status(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	conv, err := h.svc.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": err.Error()}})
		return
	}
	if conv == nil || conv.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "conversion not found"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": conv})
}

func (h *ConversionHandler) Download(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	reader, conv, err := h.svc.GetDownload(id)
	if err != nil || conv == nil || conv.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "conversion not found or not ready"}})
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "attachment; filename="+id+".svg")
	c.Header("Content-Type", "image/svg+xml")
	c.DataFromReader(http.StatusOK, -1, "image/svg+xml", reader, nil)
}
