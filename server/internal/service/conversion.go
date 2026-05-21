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
	MaxDailyConversions = 20
	BucketOriginals     = "originals"
	BucketResults       = "results"
)

type ConversionPayload struct {
	ConversionID string `json:"conversion_id"`
	OriginalKey  string `json:"original_key"`
	FormatIn     string `json:"format_in"`
}

type ConversionService struct {
	cfg     *config.Config
	repo    *repo.ConversionRepo
	storage *Storage
	client  *asynq.Client
}

func NewConversionService(cfg *config.Config, r *repo.ConversionRepo, s *Storage, c *asynq.Client) *ConversionService {
	return &ConversionService{cfg: cfg, repo: r, storage: s, client: c}
}

func (s *ConversionService) Enqueue(userID string, file io.Reader, filename string, size int64) (*model.Conversion, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".png"
	}
	formatIn := ext[1:]
	if formatIn == "jpeg" {
		formatIn = "jpg"
	}

	originalKey := fmt.Sprintf("%s/%s%s", userID, uuid.New().String(), ext)

	contentType := "image/" + formatIn
	if formatIn == "jpg" {
		contentType = "image/jpeg"
	}
	if err := s.storage.Upload(BucketOriginals, originalKey, contentType, file, size); err != nil {
		return nil, fmt.Errorf("upload: %w", err)
	}

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

	ok, err := s.repo.IncrementQuota(userID, MaxDailyConversions)
	if err != nil {
		return nil, fmt.Errorf("quota increment: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("daily quota exceeded (%d)", MaxDailyConversions)
	}

	return conv, nil
}

func (s *ConversionService) Get(id string) (*model.Conversion, error) {
	return s.repo.FindByID(id)
}

func (s *ConversionService) List(userID string, limit, offset int) ([]*model.Conversion, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.FindByUserID(userID, limit, offset)
}

func (s *ConversionService) GetDownload(id string) (io.ReadCloser, *model.Conversion, error) {
	conv, err := s.repo.FindByID(id)
	if err != nil {
		return nil, nil, err
	}
	if conv == nil || conv.Status != model.StatusCompleted || conv.SVGURL == "" {
		return nil, conv, fmt.Errorf("conversion not ready")
	}
	reader, err := s.storage.Download(BucketResults, conv.SVGURL)
	return reader, conv, err
}
