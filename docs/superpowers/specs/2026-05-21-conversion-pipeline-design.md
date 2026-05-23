# 转换管线设计

## 概述

添加核心转换能力：已认证用户上传位图图片，后台 worker 通过外部 CLI 工具将其转换为 SVG，用户可以列表/查看/下载结果。

## 组件

### 1. 转换仓库（`repo/conversion.go`）
- Create、FindByID、FindByUserID、UpdateStatus
- 纯 CRUD 操作，无业务逻辑

### 2. 存储服务（`service/storage.go`）
- Upload(file, bucket, key) → 存储到 MinIO
- Download(bucket, key) → 返回 io.Reader
- 存储桶：`originals`（原始文件）、`results`（结果文件）、`thumbnails`（缩略图）

### 3. 转换服务（`service/conversion.go`）
- Enqueue(userID, file) → 检查配额，将原始文件存储到 MinIO，插入转换记录，推送 asynq 任务
- GetStatus(id) → 返回转换记录
- ListUserConversions(userID) → 分页列表
- CheckQuota(userID) → 验证每日限额（每天 20 次）

### 4. 转换处理器（`handler/conversion.go`）
- POST /api/v1/conversions — multipart 上传，需要认证
- GET /api/v1/conversions — 列表，需要认证
- GET /api/v1/conversions/:id — 状态查询，需要认证
- GET /api/v1/conversions/:id/download — 重定向到 MinIO 预签名 URL 或流式传输

### 5. Worker（`cmd/worker/main.go`）
- asynq 服务器处理转换任务
- 从 MinIO 下载原始文件 → 运行 potrace/vtrace → 上传结果 → 更新状态
- 错误处理：将转换标记为失败并记录错误信息

### 6. 路由
- conversion 路由组挂载在 /api/v1 下，全部需要 JWT 认证

## 数据流

客户端 → 上传 → API → MinIO（原始文件）+ 数据库（pending）→ Redis（asynq）
Worker → MinIO（原始文件）→ CLI 工具 → MinIO（结果文件）→ 数据库（completed）
客户端 → GET 状态/下载 → API → MinIO（结果文件）

## 依赖
- `github.com/minio/minio-go/v7` — MinIO 客户端
- `github.com/hibiken/asynq` — 基于 Redis 的任务队列
- CLI 工具：`potrace` 或 `vtracer`，安装在 worker 容器中

## 错误处理
- 上传失败 → 4xx 响应
- 配额超限 → 429
- 转换失败 → status=failed，附带 error_message
- Worker 重试：通过 asynq 进行 3 次指数退避重试
