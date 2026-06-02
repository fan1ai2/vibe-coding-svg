package handler

import (
	"net/http"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)

// AiGenerator abstracts the AI service for testability.
type AiGenerator interface {
	Generate(userID string, req service.GenerateRequest) (*service.GenerateResponse, error)
	Quota(userID string) service.QuotaResponse
}

type AiHandler struct {
	aiService AiGenerator
}

func NewAiHandler(aiService AiGenerator) *AiHandler {
	return &AiHandler{aiService: aiService}
}

func (h *AiHandler) Generate(c *gin.Context) {
	userID := c.GetString("user_id")

	var req service.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_PARAMS", "message": "请输入图标描述（1-200字）"},
		})
		return
	}
	if req.Style == "" {
		req.Style = "line"
	}
	if req.Style != "line" && req.Style != "filled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_PARAMS", "message": "风格必须是 line 或 filled"},
		})
		return
	}
	req.Prompt = strings.TrimSpace(req.Prompt)

	resp, err := h.aiService.Generate(userID, req)
	if err != nil {
		msg := err.Error()
		code := "GENERATE_FAILED"
		status := http.StatusInternalServerError
		if strings.Contains(msg, "quota exceeded") {
			code = "QUOTA_EXCEEDED"
			status = http.StatusTooManyRequests
		}
		c.JSON(status, gin.H{
			"error": gin.H{"code": code, "message": msg},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *AiHandler) Quota(c *gin.Context) {
	userID := c.GetString("user_id")
	resp := h.aiService.Quota(userID)
	c.JSON(http.StatusOK, gin.H{"data": resp})
}
