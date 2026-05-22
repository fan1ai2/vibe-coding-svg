# Conversion Pipeline Design

## Overview

Add the core conversion capability: authenticated users upload bitmap images, a background worker converts them to SVG via external CLI tools, and users can list/check/download results.

## Components

### 1. Conversion Repo (`repo/conversion.go`)
- Create, FindByID, FindByUserID, UpdateStatus
- Pure CRUD, no business logic

### 2. Storage Service (`service/storage.go`)
- Upload(file, bucket, key) → stores to MinIO
- Download(bucket, key) → returns io.Reader
- Buckets: `originals`, `results`, `thumbnails`

### 3. Conversion Service (`service/conversion.go`)
- Enqueue(userID, file) → check quota, store original in MinIO, insert conversion row, push asynq task
- GetStatus(id) → return conversion record
- ListUserConversions(userID) → paginated list
- CheckQuota(userID) → validate daily limit (20/day)

### 4. Conversion Handlers (`handler/conversion.go`)
- POST /api/v1/conversions — multipart upload, authenticated
- GET /api/v1/conversions — list, authenticated
- GET /api/v1/conversions/:id — status, authenticated
- GET /api/v1/conversions/:id/download — redirect to MinIO presigned URL or stream

### 5. Worker (`cmd/worker/main.go`)
- asynq server processing conversion tasks
- Download original from MinIO → run potrace/vtrace → upload result → update status
- Handle errors: mark conversion as failed with error message

### 6. Routes
- conversion group under /api/v1, all JWT-protected

## Data Flow

Client → upload → API → MinIO(originals) + DB(pending) → Redis(asynq)
Worker → MinIO(originals) → CLI tool → MinIO(results) → DB(completed)
Client → GET status/download → API → MinIO(results)

## Dependencies
- `github.com/minio/minio-go/v7` — MinIO client
- `github.com/hibiken/asynq` — Redis-based task queue
- CLI tools: `potrace` or `vtracer` installed in worker container

## Error Handling
- Upload failures → 4xx response
- Quota exceeded → 429
- Conversion failure → status=failed with error_message
- Worker retries: 3 attempts with exponential backoff via asynq
