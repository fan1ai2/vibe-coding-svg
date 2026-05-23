package handler

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)

// 允许上传的图片 MIME 类型白名单
var allowedMIMETypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
	"image/bmp":  true,
	"image/tiff": true,
}

// 允许的文件扩展名白名单
var allowedExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".bmp":  true,
	".tiff": true,
}

// ConversionHandler 转换相关接口处理器
type ConversionHandler struct {
	cfg *config.Config
	svc *service.ConversionService
}

// NewConversionHandler 创建转换处理器实例
func NewConversionHandler(cfg *config.Config, svc *service.ConversionService) *ConversionHandler {
	return &ConversionHandler{cfg: cfg, svc: svc}
}

// Upload godoc
// @Summary      上传图片进行转换
// @Tags         conversions
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Param        file  formData  file  true  "PNG 或 JPEG 图片文件"
// @Success      201   {object}  object{data=model.Conversion}
// @Failure      400   {object}  object{error=object{code=string,message=string}}
// @Failure      413   {object}  object{error=object{code=string,message=string}}
// @Failure      429   {object}  object{error=object{code=string,message=string}}
// @Router       /conversions [post]
func (h *ConversionHandler) Upload(c *gin.Context) {
	userID := c.GetString("user_id")

	// 从表单中获取上传文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "NO_FILE", "message": "请上传文件"}})
		return
	}
	defer file.Close()

	// 检查文件大小是否超限
	if header.Size > h.cfg.MaxFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"code": "FILE_TOO_LARGE", "message": "文件大小超出限制"}})
		return
	}

	// 校验文件扩展名
	dotIdx := strings.LastIndex(header.Filename, ".")
	if dotIdx < 0 || !allowedExtensions[strings.ToLower(header.Filename[dotIdx:])] {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_TYPE", "message": "不支持的文件类型，仅支持 PNG、JPEG、WebP、BMP、TIFF"}})
		return
	}

	// 校验 MIME 类型（读取文件前 512 字节检测真实类型）
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	detectedType := http.DetectContentType(buf[:n])
	if !allowedMIMETypes[detectedType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_TYPE", "message": "文件内容不是支持的图片格式"}})
		return
	}

	// 将转换任务加入队列（用 MultiReader 合并已读取的头部和剩余数据）
	combinedReader := io.MultiReader(bytes.NewReader(buf[:n]), file)
	conv, err := h.svc.Enqueue(userID, combinedReader, header.Filename, header.Size)
	if err != nil {
		if strings.Contains(err.Error(), "quota") {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "QUOTA_EXCEEDED", "message": "今日配额已用完"}})
			return
		}
		log.Printf("[ERROR] upload enqueue user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "UPLOAD_FAILED", "message": "上传失败，请稍后重试"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": conv})
}

// List godoc
// @Summary      获取转换记录列表
// @Tags         conversions
// @Security     BearerAuth
// @Param        limit   query     int  false  "每页数量"  default(20)
// @Param        offset  query     int  false  "偏移量" default(0)
// @Success      200     {object}  object{data=[]model.Conversion}
// @Router       /conversions [get]
func (h *ConversionHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	list, err := h.svc.List(userID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] list conversions user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "LIST_FAILED", "message": "获取列表失败"}})
		return
	}
	// 确保返回空数组而非 null
	if list == nil {
		list = make([]*model.Conversion, 0)
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// Status godoc
// @Summary      查询转换状态
// @Tags         conversions
// @Security     BearerAuth
// @Param        id   path      string  true  "转换 ID"
// @Success      200  {object}  object{data=model.Conversion}
// @Failure      404  {object}  object{error=object{code=string,message=string}}
// @Router       /conversions/{id} [get]
func (h *ConversionHandler) Status(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	conv, err := h.svc.Get(id)
	if err != nil {
		log.Printf("[ERROR] get conversion id=%s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": "查询失败"}})
		return
	}
	// 验证记录存在且属于当前用户
	if conv == nil || conv.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "转换记录不存在"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": conv})
}

// Download godoc
// @Summary      下载 SVG 结果文件
// @Tags         conversions
// @Security     BearerAuth
// @Param        id   path      string  true  "转换 ID"
// @Success      200  {file}    image/svg+xml
// @Failure      404  {object}  object{error=object{code=string,message=string}}
// @Router       /conversions/{id}/download [get]
func (h *ConversionHandler) Download(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	// 从对象存储获取 SVG 文件
	reader, conv, err := h.svc.GetDownload(id)
	if err != nil || conv == nil || conv.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "转换记录不存在或尚未完成"}})
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "attachment; filename="+id+".svg")
	c.Header("Content-Type", "image/svg+xml")
	c.DataFromReader(http.StatusOK, -1, "image/svg+xml", reader, nil)
}
