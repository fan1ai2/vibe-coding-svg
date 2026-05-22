# Conversion Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the complete SVG conversion pipeline: upload, background processing via asynq, status tracking, and download.

**Architecture:** API server handles uploads and stores originals in MinIO, then enqueues asynq tasks to Redis. A separate worker process picks up tasks, runs vtracer CLI for bitmap-to-SVG conversion, uploads results to MinIO, and updates conversion status. Quota is enforced per user per day (20 conversions).

**Tech Stack:** Go 1.25, Gin, minio-go v7, asynq, vtracer CLI, PostgreSQL

**File Map:**
- Create: `server/internal/repo/conversion.go` — Conversion + quota DB ops
- Create: `server/internal/service/storage.go` — MinIO upload/download wrapper
- Create: `server/internal/service/conversion.go` — Enqueue/status/list biz logic
- Create: `server/internal/handler/conversion.go` — HTTP handlers
- Create: `server/internal/worker/converter.go` — CLI wrapper for vtracer
- Create: `server/internal/worker/worker.go` — asynq task handler
- Modify: `server/internal/router/router.go` — Add conversion routes
- Modify: `server/cmd/worker/main.go` — Wire up asynq server
- Modify: `server/go.mod` / `server/go.sum` — Add minio-go, asynq, uuid

---

### Task 1: Add Dependencies

**Files:** Modify `server/go.mod`, `server/go.sum`

- [ ] **Step 1: Add minio-go, asynq, uuid**

```bash
cd /svg-project/server && go get github.com/minio/minio-go/v7 github.com/hibiken/asynq github.com/google/uuid
```

Expected: downloads modules, updates go.mod and go.sum.

- [ ] **Step 2: Verify go.mod has new deps**

```bash
cd /svg-project/server && grep -E "minio-go|asynq|uuid" go.mod
```

Expected: three require lines for the new modules.

---

### Task 2: Conversion Repo

**Files:** Create `server/internal/repo/conversion.go`

- [ ] **Step 1: Write ConversionRepo**

```go
package repo

import (
	"database/sql"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
)

type ConversionRepo struct{ db *sql.DB }

func NewConversionRepo(db *sql.DB) *ConversionRepo { return &ConversionRepo{db} }

func (r *ConversionRepo) Create(c *model.Conversion) error {
	return r.db.QueryRow(
		`INSERT INTO conversions (user_id, status, original_url, format_in, file_size_in)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
		c.UserID, c.Status, c.OriginalURL, c.FormatIn, c.FileSizeIn,
	).Scan(&c.ID, &c.CreatedAt)
}

func (r *ConversionRepo) FindByID(id string) (*model.Conversion, error) {
	c := &model.Conversion{}
	err := r.db.QueryRow(
		`SELECT id, user_id, status, original_url, svg_url, thumbnail_url,
		 file_size_in, file_size_out, path_count, color_count, format_in,
		 error_message, created_at, completed_at FROM conversions WHERE id=$1`, id,
	).Scan(&c.ID, &c.UserID, &c.Status, &c.OriginalURL, &c.SVGURL, &c.ThumbnailURL,
		&c.FileSizeIn, &c.FileSizeOut, &c.PathCount, &c.ColorCount, &c.FormatIn,
		&c.ErrorMessage, &c.CreatedAt, &c.CompletedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *ConversionRepo) FindByUserID(userID string, limit, offset int) ([]*model.Conversion, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, status, original_url, svg_url, thumbnail_url,
		 file_size_in, file_size_out, path_count, color_count, format_in,
		 error_message, created_at, completed_at
		 FROM conversions WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.Conversion
	for rows.Next() {
		c := &model.Conversion{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.Status, &c.OriginalURL, &c.SVGURL, &c.ThumbnailURL,
			&c.FileSizeIn, &c.FileSizeOut, &c.PathCount, &c.ColorCount, &c.FormatIn,
			&c.ErrorMessage, &c.CreatedAt, &c.CompletedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *ConversionRepo) UpdateStatus(id, status, errMsg string) error {
	var completedAt *time.Time
	if status == model.StatusCompleted || status == model.StatusFailed {
		now := time.Now()
		completedAt = &now
	}
	_, err := r.db.Exec(
		`UPDATE conversions SET status=$1, error_message=$2, completed_at=$3 WHERE id=$4`,
		status, errMsg, completedAt, id,
	)
	return err
}

func (r *ConversionRepo) UpdateResult(id, svgURL, thumbnailURL string, fileSizeOut, pathCount, colorCount int64) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE conversions SET status=$1, svg_url=$2, thumbnail_url=$3,
		 file_size_out=$4, path_count=$5, color_count=$6, completed_at=$7 WHERE id=$8`,
		model.StatusCompleted, svgURL, thumbnailURL, fileSizeOut, pathCount, colorCount, now, id,
	)
	return err
}

func (r *ConversionRepo) GetTodayQuota(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT count FROM daily_quotas WHERE user_id=$1 AND date=CURRENT_DATE`, userID,
	).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

func (r *ConversionRepo) IncrementQuota(userID string) error {
	_, err := r.db.Exec(
		`INSERT INTO daily_quotas (user_id, date, count) VALUES ($1, CURRENT_DATE, 1)
		 ON CONFLICT (user_id, date) DO UPDATE SET count = daily_quotas.count + 1`,
		userID,
	)
	return err
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /svg-project/server && go build ./internal/repo/...
```

Expected: compiles without errors.

---

### Task 3: Storage Service (MinIO)

**Files:** Create `server/internal/service/storage.go`

- [ ] **Step 1: Write Storage service**

```go
package service

import (
	"context"
	"io"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client *minio.Client
}

func NewStorage(cfg *config.Config) (*Storage, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	return &Storage{client: client}, nil
}

func (s *Storage) EnsureBucket(name string) error {
	exists, err := s.client.BucketExists(context.Background(), name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.client.MakeBucket(context.Background(), name, minio.MakeBucketOptions{})
}

func (s *Storage) Upload(bucket, key, contentType string, reader io.Reader, size int64) error {
	_, err := s.client.PutObject(context.Background(), bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *Storage) Download(bucket, key string) (io.ReadCloser, error) {
	return s.client.GetObject(context.Background(), bucket, key, minio.GetObjectOptions{})
}

func (s *Storage) PresignedGetURL(bucket, key string, expiry time.Duration) (string, error) {
	u, err := s.client.PresignedGet(context.Background(), bucket, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /svg-project/server && go build ./internal/service/...
```

Expected: compiles without errors.

---

### Task 4: Conversion Service

**Files:** Create `server/internal/service/conversion.go`

- [ ] **Step 1: Write ConversionService**

```go
package service

import (
	"encoding/json"
	"fmt"
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
	count, err := s.repo.GetTodayQuota(userID)
	if err != nil {
		return nil, fmt.Errorf("quota check: %w", err)
	}
	if count >= MaxDailyConversions {
		return nil, fmt.Errorf("daily quota exceeded (%d/%d)", count, MaxDailyConversions)
	}

	ext := filepath.Ext(filename)
	formatIn := ext[1:]
	if formatIn == "jpeg" {
		formatIn = "jpg"
	}

	originalKey := fmt.Sprintf("%s/%s%s", userID, uuid.New().String(), ext)

	if err := s.storage.Upload(BucketOriginals, originalKey, "image/"+formatIn, file, size); err != nil {
		return nil, fmt.Errorf("upload: %w", err)
	}

	conv := &model.Conversion{
		UserID:      userID,
		Status:      model.StatusPending,
		OriginalURL: originalKey,
		FormatIn:    formatIn,
		FileSizeIn:  size,
	}
	if err := s.repo.Create(conv); err != nil {
		return nil, fmt.Errorf("create conversion: %w", err)
	}

	if err := s.repo.IncrementQuota(userID); err != nil {
		return nil, fmt.Errorf("quota increment: %w", err)
	}

	payload := ConversionPayload{
		ConversionID: conv.ID,
		OriginalKey:  originalKey,
		FormatIn:     formatIn,
	}
	body, _ := json.Marshal(payload)
	task := asynq.NewTask("conversion:process", body)
	if _, err := s.client.Enqueue(task); err != nil {
		return nil, fmt.Errorf("enqueue: %w", err)
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
```

Missing import: add `"io"` at the top. Full imports block:

```go
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /svg-project/server && go build ./internal/service/...
```

Expected: compiles without errors.

---

### Task 5: Conversion Handlers

**Files:** Create `server/internal/handler/conversion.go`

- [ ] **Step 1: Write ConversionHandler**

```go
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
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
```

Note: the List handler references `[]*model.Conversion` in the nil check. Add the model import:

```go
import (
	"net/http"
	"strconv"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)
```

- [ ] **Step 2: Verify compilation**

```bash
cd /svg-project/server && go build ./internal/handler/...
```

Expected: compiles without errors.

---

### Task 6: Update Router

**Files:** Modify `server/internal/router/router.go`

- [ ] **Step 1: Add conversion routes**

Replace the entire file. Add new imports for repo, asynq, and use the new services:

```go
package router

import (
	"database/sql"
	"log"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/handler"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/middleware"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

func Setup(cfg *config.Config, db *sql.DB) *gin.Engine {
	r := gin.Default()

	userRepo := repo.NewUserRepo(db)
	authSvc := service.NewAuthService(cfg, userRepo)
	authH := handler.NewAuthHandler(cfg, authSvc)

	storage, err := service.NewStorage(cfg)
	if err != nil {
		log.Fatalf("storage init: %v", err)
	}
	if err := storage.EnsureBucket(service.BucketOriginals); err != nil {
		log.Fatalf("bucket originals: %v", err)
	}
	if err := storage.EnsureBucket(service.BucketResults); err != nil {
		log.Fatalf("bucket results: %v", err)
	}

	convRepo := repo.NewConversionRepo(db)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr})
	convSvc := service.NewConversionService(cfg, convRepo, storage, asynqClient)
	convH := handler.NewConversionHandler(cfg, convSvc)

	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(100))

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.GET("/github/login", authH.GithubLogin)
			auth.GET("/github/callback", authH.GithubCallback)
			auth.GET("/google/login", authH.GoogleLogin)
			auth.GET("/google/callback", authH.GoogleCallback)
			auth.POST("/refresh", middleware.JWTAuth(cfg), authH.Refresh)
			auth.GET("/me", middleware.JWTAuth(cfg), authH.Me)
		}

		conversions := api.Group("/conversions")
		conversions.Use(middleware.JWTAuth(cfg))
		{
			conversions.POST("", convH.Upload)
			conversions.GET("", convH.List)
			conversions.GET("/:id", convH.Status)
			conversions.GET("/:id/download", convH.Download)
		}
	}

	return r
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /svg-project/server && go build ./internal/router/...
```

Expected: compiles without errors.

---

### Task 7: Worker Converter (CLI wrapper)

**Files:** Create `server/internal/worker/converter.go`

- [ ] **Step 1: Write vtracer CLI wrapper**

```go
package worker

import (
	"fmt"
	"os/exec"
	"strings"
)

func ConvertRasterToSVG(inputPath, outputPath string) error {
	cmd := exec.Command("vtracer", "--input", inputPath, "--output", outputPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("vtracer: %s: %w", string(out), err)
	}
	return nil
}

func CountSVGPaths(data []byte) int {
	return strings.Count(string(data), "<path ")
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /svg-project/server && mkdir -p internal/worker && go build ./internal/worker/...
```

Expected: compiles without errors.

---

### Task 8: Worker Task Handler

**Files:** Create `server/internal/worker/worker.go`

- [ ] **Step 1: Write asynq task handler**

```go
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/hibiken/asynq"
)

type ConversionWorker struct {
	cfg     *config.Config
	repo    *repo.ConversionRepo
	storage *service.Storage
}

func NewConversionWorker(cfg *config.Config, r *repo.ConversionRepo, s *service.Storage) *ConversionWorker {
	return &ConversionWorker{cfg: cfg, repo: r, storage: s}
}

func (w *ConversionWorker) HandleProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload service.ConversionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if err := w.repo.UpdateStatus(payload.ConversionID, model.StatusProcessing, ""); err != nil {
		return fmt.Errorf("update status to processing: %w", err)
	}

	reader, err := w.storage.Download(service.BucketOriginals, payload.OriginalKey)
	if err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "download failed: "+err.Error())
		return fmt.Errorf("download original: %w", err)
	}
	defer reader.Close()

	tmpDir := os.TempDir()
	inPath := filepath.Join(tmpDir, payload.ConversionID+"_in."+payload.FormatIn)
	outPath := filepath.Join(tmpDir, payload.ConversionID+"_out.svg")

	inFile, err := os.Create(inPath)
	if err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "temp file: "+err.Error())
		return fmt.Errorf("create temp input: %w", err)
	}
	if _, err := io.Copy(inFile, reader); err != nil {
		inFile.Close()
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "write temp: "+err.Error())
		return fmt.Errorf("write temp input: %w", err)
	}
	inFile.Close()
	defer os.Remove(inPath)
	defer os.Remove(outPath)

	if err := ConvertRasterToSVG(inPath, outPath); err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "conversion failed: "+err.Error())
		return fmt.Errorf("convert: %w", err)
	}

	svgData, err := os.ReadFile(outPath)
	if err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "read result: "+err.Error())
		return fmt.Errorf("read svg result: %w", err)
	}

	resultKey := payload.OriginalKey + ".svg"
	resultFile, err := os.Open(outPath)
	if err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "open result: "+err.Error())
		return fmt.Errorf("open svg file: %w", err)
	}
	defer resultFile.Close()

	fi, _ := resultFile.Stat()
	if err := w.storage.Upload(service.BucketResults, resultKey, "image/svg+xml", resultFile, fi.Size()); err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "upload result: "+err.Error())
		return fmt.Errorf("upload svg result: %w", err)
	}

	pathCount := CountSVGPaths(svgData)
	fileSizeOut := int64(len(svgData))

	if err := w.repo.UpdateResult(payload.ConversionID, resultKey, "", fileSizeOut, pathCount, 0); err != nil {
		return fmt.Errorf("update result in db: %w", err)
	}

	return nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /svg-project/server && go build ./internal/worker/...
```

Expected: compiles without errors.

---

### Task 9: Worker Main

**Files:** Modify `server/cmd/worker/main.go`

- [ ] **Step 1: Rewrite worker main.go**

```go
package main

import (
	"database/sql"
	"log"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/worker"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	log.Println("connected to postgres")

	storage, err := service.NewStorage(cfg)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	if err := storage.EnsureBucket(service.BucketOriginals); err != nil {
		log.Fatalf("bucket originals: %v", err)
	}
	if err := storage.EnsureBucket(service.BucketResults); err != nil {
		log.Fatalf("bucket results: %v", err)
	}

	convRepo := repo.NewConversionRepo(db)
	convWorker := worker.NewConversionWorker(cfg, convRepo, storage)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{Concurrency: 4},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("conversion:process", convWorker.HandleProcessTask)

	log.Printf("Worker starting, redis=%s", cfg.RedisAddr)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("worker: %v", err)
	}
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /svg-project/server && go build ./cmd/worker/...
```

Expected: compiles without errors.

---

### Task 10: Full Build Verification

- [ ] **Step 1: Build both binaries**

```bash
cd /svg-project/server && go build -o /dev/null ./cmd/api/... && go build -o /dev/null ./cmd/worker/...
```

Expected: both compile successfully, no output on success.

- [ ] **Step 2: Run go vet**

```bash
cd /svg-project/server && go vet ./...
```

Expected: no warnings or errors.

- [ ] **Step 3: Commit**

```bash
cd /svg-project && git add server/go.mod server/go.sum \
  server/internal/repo/conversion.go \
  server/internal/service/storage.go \
  server/internal/service/conversion.go \
  server/internal/handler/conversion.go \
  server/internal/worker/ \
  server/internal/router/router.go \
  server/cmd/worker/main.go
git commit -m "feat: add conversion pipeline with MinIO, asynq, and vtracer worker"
```
