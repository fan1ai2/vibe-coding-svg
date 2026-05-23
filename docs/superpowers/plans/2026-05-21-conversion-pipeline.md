# 转换管线实现计划

> **面向自动化工作者:** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 按任务逐步实现此计划。步骤使用复选框 (`- [ ]`) 语法进行跟踪。

**目标：** 构建完整的 SVG 转换管线：上传、通过 asynq 进行后台处理、状态跟踪和下载。

**架构：** API 服务器处理上传并将原始文件存储到 MinIO，然后将 asynq 任务入队到 Redis。一个独立的 worker 进程拾取任务，运行 vtracer CLI 进行位图转 SVG 的转换，将结果上传到 MinIO，并更新转换状态。每个用户每天有配额限制（20 次转换）。

**技术栈：** Go 1.25、Gin、minio-go v7、asynq、vtracer CLI、PostgreSQL

**文件清单：**
- 新建：`server/internal/repo/conversion.go` — 转换记录和配额的数据库操作
- 新建：`server/internal/service/storage.go` — MinIO 上传/下载封装
- 新建：`server/internal/service/conversion.go` — 入队/状态/列表业务逻辑
- 新建：`server/internal/handler/conversion.go` — HTTP 处理器
- 新建：`server/internal/worker/converter.go` — vtracer CLI 封装
- 新建：`server/internal/worker/worker.go` — asynq 任务处理器
- 修改：`server/internal/router/router.go` — 添加转换路由
- 修改：`server/cmd/worker/main.go` — 连接 asynq 服务器
- 修改：`server/go.mod` / `server/go.sum` — 添加 minio-go、asynq、uuid

---

### 任务 1：添加依赖

**涉及文件：** 修改 `server/go.mod`、`server/go.sum`

- [ ] **步骤 1：添加 minio-go、asynq、uuid**

```bash
cd /svg-project/server && go get github.com/minio/minio-go/v7 github.com/hibiken/asynq github.com/google/uuid
```

预期结果：下载模块，更新 go.mod 和 go.sum。

- [ ] **步骤 2：验证 go.mod 包含新依赖**

```bash
cd /svg-project/server && grep -E "minio-go|asynq|uuid" go.mod
```

预期结果：三个新模块的 require 行。

---

### 任务 2：转换记录仓库

**涉及文件：** 新建 `server/internal/repo/conversion.go`

- [ ] **步骤 1：编写 ConversionRepo**

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

- [ ] **步骤 2：验证编译**

```bash
cd /svg-project/server && go build ./internal/repo/...
```

预期结果：编译通过，无错误。

---

### 任务 3：存储服务（MinIO）

**涉及文件：** 新建 `server/internal/service/storage.go`

- [ ] **步骤 1：编写 Storage 服务**

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

- [ ] **步骤 2：验证编译**

```bash
cd /svg-project/server && go build ./internal/service/...
```

预期结果：编译通过，无错误。

---

### 任务 4：转换服务

**涉及文件：** 新建 `server/internal/service/conversion.go`

- [ ] **步骤 1：编写 ConversionService**

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
	MaxDailyConversions = 20    // 每日最大转换次数
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

缺少导入：在顶部添加 `"io"`。完整导入块：

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

- [ ] **步骤 2：验证编译**

```bash
cd /svg-project/server && go build ./internal/service/...
```

预期结果：编译通过，无错误。

---

### 任务 5：转换处理器

**涉及文件：** 新建 `server/internal/handler/conversion.go`

- [ ] **步骤 1：编写 ConversionHandler**

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

注意：List 处理器在 nil 检查中引用了 `[]*model.Conversion`。添加 model 导入：

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

- [ ] **步骤 2：验证编译**

```bash
cd /svg-project/server && go build ./internal/handler/...
```

预期结果：编译通过，无错误。

---

### 任务 6：更新路由

**涉及文件：** 修改 `server/internal/router/router.go`

- [ ] **步骤 1：添加转换路由**

替换整个文件。添加 repo、asynq 的新导入，并使用新服务：

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

- [ ] **步骤 2：验证编译**

```bash
cd /svg-project/server && go build ./internal/router/...
```

预期结果：编译通过，无错误。

---

### 任务 7：Worker 转换器（CLI 封装）

**涉及文件：** 新建 `server/internal/worker/converter.go`

- [ ] **步骤 1：编写 vtracer CLI 封装**

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

- [ ] **步骤 2：验证编译**

```bash
cd /svg-project/server && mkdir -p internal/worker && go build ./internal/worker/...
```

预期结果：编译通过，无错误。

---

### 任务 8：Worker 任务处理器

**涉及文件：** 新建 `server/internal/worker/worker.go`

- [ ] **步骤 1：编写 asynq 任务处理器**

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

	// 更新状态为处理中
	if err := w.repo.UpdateStatus(payload.ConversionID, model.StatusProcessing, ""); err != nil {
		return fmt.Errorf("update status to processing: %w", err)
	}

	// 从对象存储下载原始文件
	reader, err := w.storage.Download(service.BucketOriginals, payload.OriginalKey)
	if err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "download failed: "+err.Error())
		return fmt.Errorf("download original: %w", err)
	}
	defer reader.Close()

	// 创建临时文件
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

	// 执行转换
	if err := ConvertRasterToSVG(inPath, outPath); err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "conversion failed: "+err.Error())
		return fmt.Errorf("convert: %w", err)
	}

	// 读取 SVG 结果
	svgData, err := os.ReadFile(outPath)
	if err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "read result: "+err.Error())
		return fmt.Errorf("read svg result: %w", err)
	}

	// 上传结果到对象存储
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

	// 统计并更新结果
	pathCount := CountSVGPaths(svgData)
	fileSizeOut := int64(len(svgData))

	if err := w.repo.UpdateResult(payload.ConversionID, resultKey, "", fileSizeOut, pathCount, 0); err != nil {
		return fmt.Errorf("update result in db: %w", err)
	}

	return nil
}
```

- [ ] **步骤 2：验证编译**

```bash
cd /svg-project/server && go build ./internal/worker/...
```

预期结果：编译通过，无错误。

---

### 任务 9：Worker 主程序

**涉及文件：** 修改 `server/cmd/worker/main.go`

- [ ] **步骤 1：重写 worker main.go**

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

	// 连接数据库
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	log.Println("connected to postgres")

	// 初始化对象存储
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

	// 启动 asynq 服务器
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

- [ ] **步骤 2：验证编译**

```bash
cd /svg-project/server && go build ./cmd/worker/...
```

预期结果：编译通过，无错误。

---

### 任务 10：完整构建验证

- [ ] **步骤 1：构建两个二进制文件**

```bash
cd /svg-project/server && go build -o /dev/null ./cmd/api/... && go build -o /dev/null ./cmd/worker/...
```

预期结果：两个都编译成功，成功时无输出。

- [ ] **步骤 2：运行 go vet**

```bash
cd /svg-project/server && go vet ./...
```

预期结果：无警告或错误。

- [ ] **步骤 3：提交**

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
