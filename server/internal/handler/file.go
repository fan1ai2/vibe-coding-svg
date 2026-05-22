package handler

import (
	"net/http"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	storage *service.Storage
}

func NewFileHandler(storage *service.Storage) *FileHandler {
	return &FileHandler{storage: storage}
}

func (h *FileHandler) Serve(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	reader, err := h.storage.Download(bucket, key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "file not found"}})
		return
	}
	defer reader.Close()

	contentType := "application/octet-stream"
	if strings.HasSuffix(key, ".svg") {
		contentType = "image/svg+xml"
	} else if strings.HasSuffix(key, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(key, ".jpg") || strings.HasSuffix(key, ".jpeg") {
		contentType = "image/jpeg"
	}

	c.DataFromReader(http.StatusOK, -1, contentType, reader, nil)
}
