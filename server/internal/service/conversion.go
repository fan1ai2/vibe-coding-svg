package service

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	MaxDailyConversions = 20   // 每日最大转换次数
	BucketOriginals     = "originals" // 原始文件存储桶
	BucketResults       = "results"   // 转换结果存储桶
)

// ConversionPayload 发送到 asynq 任务队列的转换任务载荷
type ConversionPayload struct {
	ConversionID string `json:"conversion_id"` // 转换任务 ID
	OriginalKey  string `json:"original_key"`  // 原始文件在对象存储中的路径
	FormatIn     string `json:"format_in"`     // 输入文件格式
}

// ConversionService 转换服务，负责文件上传、任务入队和结果查询
type ConversionService struct {
	cfg     *config.Config
	repo    *repo.ConversionRepo
	storage *Storage
	client  *asynq.Client // asynq 任务队列客户端
}

// NewConversionService 创建转换服务实例
func NewConversionService(cfg *config.Config, r *repo.ConversionRepo, s *Storage, c *asynq.Client) *ConversionService {
	return &ConversionService{cfg: cfg, repo: r, storage: s, client: c}
}

// Enqueue 上传原始文件到对象存储，创建转换记录并将任务加入队列
func (s *ConversionService) Enqueue(userID string, file io.Reader, filename string, size int64) (*model.Conversion, error) {
	// Guest quota: lifetime 3 conversions max
	provider, err := s.repo.FindProviderByID(userID)
	if err != nil {
		return nil, fmt.Errorf("quota check: %w", err)
	}
	if provider == "guest" {
		count, err := s.repo.CountByUserID(userID)
		if err != nil {
			return nil, fmt.Errorf("quota count: %w", err)
		}
		if count >= 3 {
			return nil, fmt.Errorf("试用次数已用完（%d/3），请登录后继续使用", count)
		}
	}

	// 解析文件扩展名确定输入格式
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".png"
	}
	formatIn := ext[1:]
	if formatIn == "jpeg" {
		formatIn = "jpg"
	}

	// 生成唯一存储路径并上传到对象存储
	originalKey := fmt.Sprintf("%s/%s%s", userID, uuid.New().String(), ext)

	contentType := "image/" + formatIn
	if formatIn == "jpg" {
		contentType = "image/jpeg"
	}
	if err := s.storage.Upload(BucketOriginals, originalKey, contentType, file, size); err != nil {
		return nil, fmt.Errorf("upload: %w", err)
	}

	// 构建任务载荷并推送到 asynq 队列
	convID := uuid.New().String()
	payload := ConversionPayload{
		ConversionID: convID,
		OriginalKey:  originalKey,
		FormatIn:     formatIn,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	task := asynq.NewTask("conversion:process", body)
	if _, err := s.client.Enqueue(task); err != nil {
		return nil, fmt.Errorf("enqueue: %w", err)
	}

	// 创建数据库转换记录
	conv := &model.Conversion{
		ID:          convID,
		UserID:      userID,
		Status:      model.StatusPending,
		OriginalURL: originalKey,
		FormatIn:    formatIn,
		FileSizeIn:  size,
	}
	if err := s.repo.Create(conv); err != nil {
		return nil, fmt.Errorf("create conversion: %w", err)
	}

	// 检查并递增每日配额
	ok, err := s.repo.IncrementQuota(userID, MaxDailyConversions)
	if err != nil {
		return nil, fmt.Errorf("quota increment: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("每日配额已用完 (%d)", MaxDailyConversions)
	}

	return conv, nil
}

// Get 根据 ID 查询单条转换记录
func (s *ConversionService) Get(id string) (*model.Conversion, error) {
	return s.repo.FindByID(id)
}

// List 查询用户的转换记录列表，支持分页
func (s *ConversionService) List(userID string, limit, offset int) ([]*model.Conversion, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.FindByUserID(userID, limit, offset)
}

// GetDownload 获取已完成转换的 SVG 文件下载流
func (s *ConversionService) GetDownload(id string) (io.ReadCloser, *model.Conversion, error) {
	conv, err := s.repo.FindByID(id)
	if err != nil {
		return nil, nil, err
	}
	// 只有状态为已完成且有 SVG 文件路径的记录才能下载
	if conv == nil || conv.Status != model.StatusCompleted || conv.SVGURL == nil || *conv.SVGURL == "" {
		return nil, conv, fmt.Errorf("conversion not ready")
	}
	reader, err := s.storage.Download(BucketResults, *conv.SVGURL)
	return reader, conv, err
}
