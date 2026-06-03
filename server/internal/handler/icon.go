package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)

type IconHandler struct {
	svc *service.IconService
}

func NewIconHandler(svc *service.IconService) *IconHandler {
	return &IconHandler{svc: svc}
}

type createIconRequest struct {
	Name       string               `json:"name" binding:"required"`
	SvgContent string               `json:"svg_content" binding:"required"`
	IsPublic   bool                 `json:"is_public"`
	Tags       []service.TagInput   `json:"tags"`
	Theme      string               `json:"theme"`
}

func (h *IconHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	var req createIconRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_BODY", "message": "请提供 name 和 svg_content"}})
		return
	}
	if len(req.SvgContent) > 2*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"code": "TOO_LARGE", "message": "SVG 内容超过 2MB 限制"}})
		return
	}
	icon, err := h.svc.Create(userID, service.CreateIconInput{
		Name: req.Name, SvgContent: req.SvgContent, IsPublic: req.IsPublic,
		Tags: req.Tags, Theme: req.Theme,
	})
	if err != nil {
		log.Printf("[ERROR] create icon user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "CREATE_FAILED", "message": "创建图标失败"}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": icon})
}

type batchIconRequest struct {
	Icons []struct {
		Name       string             `json:"name" binding:"required"`
		SvgContent string             `json:"svg_content" binding:"required"`
		IsPublic   bool               `json:"is_public"`
		Tags       []service.TagInput `json:"tags"`
		Theme      string             `json:"theme"`
	} `json:"icons" binding:"required"`
}

func (h *IconHandler) BatchCreate(c *gin.Context) {
	userID := c.GetString("user_id")
	var req batchIconRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_BODY", "message": "请提供 icons 数组"}})
		return
	}
	if len(req.Icons) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "TOO_MANY", "message": "批量创建最多 50 个图标"}})
		return
	}
	inputs := make([]service.BatchIconInput, len(req.Icons))
	for i, ic := range req.Icons {
		inputs[i] = service.BatchIconInput{
			Name: ic.Name, SvgContent: ic.SvgContent, IsPublic: ic.IsPublic,
			Tags: ic.Tags, Theme: ic.Theme,
		}
	}
	icons, err := h.svc.BatchCreate(userID, inputs)
	if err != nil {
		log.Printf("[ERROR] batch create icon user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "CREATE_FAILED", "message": "批量创建失败"}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": icons})
}

func (h *IconHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	icons, err := h.svc.ListPublic(limit, offset)
	if err != nil {
		log.Printf("[ERROR] list icons: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "LIST_FAILED", "message": "获取列表失败"}})
		return
	}
	if icons == nil {
		icons = make([]*model.Icon, 0)
	}
	c.JSON(http.StatusOK, gin.H{"data": icons})
}

func (h *IconHandler) Get(c *gin.Context) {
	id := c.Param("id")
	icon, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": "查询失败"}})
		return
	}
	if icon == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "图标不存在"}})
		return
	}
	// If private, only owner can view
	userID := c.GetString("user_id")
	if !icon.IsPublic && (userID == "" || icon.UserID != userID) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "图标不存在"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": icon})
}

func (h *IconHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	icon, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": "查询失败"}})
		return
	}
	if icon == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "图标不存在"}})
		return
	}
	if icon.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "FORBIDDEN", "message": "无权删除此图标"}})
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		log.Printf("[ERROR] delete icon id=%s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DELETE_FAILED", "message": "删除失败"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true}})
}

func (h *IconHandler) Search(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	userID := c.GetString("user_id")

	tags := []string{}
	if t := c.Query("tags"); t != "" {
		tags = strings.Split(t, ",")
	}

	icons, err := h.svc.Search(repo.IconSearchParams{
		Query:  c.Query("q"),
		Tags:   tags,
		Color:  c.Query("color"),
		Theme:  c.Query("theme"),
		Sort:   c.DefaultQuery("sort", "newest"),
		Limit:  limit,
		Offset: offset,
		UserID: userID,
	})
	if err != nil {
		log.Printf("[ERROR] search icons: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "SEARCH_FAILED", "message": "搜索失败"}})
		return
	}
	if icons == nil {
		icons = make([]*model.Icon, 0)
	}
	c.JSON(http.StatusOK, gin.H{"data": icons})
}

func (h *IconHandler) Recommend(c *gin.Context) {
	id := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	icons, err := h.svc.Recommend(id, limit)
	if err != nil {
		log.Printf("[ERROR] recommend icon id=%s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "RECOMMEND_FAILED", "message": "推荐查询失败"}})
		return
	}
	if icons == nil {
		icons = make([]*model.Icon, 0)
	}
	c.JSON(http.StatusOK, gin.H{"data": icons})
}
