# Image-to-SVG Converter — Design Spec

**Date:** 2026-05-20
**Status:** Approved

## Product Summary

A web-based image-to-SVG conversion tool. Users upload raster images (PNG/JPG/WebP), the system converts them to SVG vector graphics using a traditional vectorization engine, and stores results in a personal library. Free tool, no marketplace in MVP.

## Tech Stack

- **Backend:** Go (gin router, JWT auth, goroutine-based worker)
- **Frontend:** React SPA (Vite, TypeScript, Tailwind CSS)
- **Database:** PostgreSQL
- **Cache/Queue:** Redis
- **Storage:** MinIO (S3-compatible)
- **Vectorization:** potrace / autotrace (wrapped as CLI call in worker)

## Architecture

Frontend-backend separation. Go API server (:8080) serves RESTful JSON. React SPA served via Nginx/CDN in production, Vite dev server in development. Go Worker runs as a separate process, consuming jobs from Redis queue.

```
Browser → React SPA → Go API (:8080) → PostgreSQL
                                        → Redis (job queue)
                                        → MinIO (files)
                    Go Worker → Redis (consume) → MinIO → Vectorize → MinIO
```

### Key Principles

- Stateless API — all state in PostgreSQL/Redis, API horizontally scalable
- Async-first — upload returns task_id immediately, worker processes in background
- External file storage — all files in MinIO, server disks are ephemeral

## Core Workflow

1. **Upload** — POST multipart/form-data to `/api/v1/conversions`, returns `{task_id, status: "pending"}`
2. **Processing** — Worker picks job from Redis queue, downloads original from MinIO, runs vectorization, uploads SVG + thumbnail to MinIO, updates DB status to `completed`
3. **Preview** — Frontend polls `GET /api/v1/conversions/:id` every 1s. When completed, shows side-by-side comparison (original vs SVG) with zoom/pan, metadata (path count, color count, size reduction)
4. **Download** — `GET /api/v1/conversions/:id/download` streams the SVG file

## Database Schema

### users
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | gen_random_uuid() |
| email | VARCHAR(255) UNIQUE | from OAuth |
| name | VARCHAR(100) | display name |
| avatar_url | VARCHAR(500) | from OAuth |
| provider | VARCHAR(20) NOT NULL | 'github' or 'google' |
| provider_id | VARCHAR(100) NOT NULL | OAuth unique ID |
| created_at | TIMESTAMPTZ | DEFAULT now() |

UNIQUE(provider, provider_id)

### conversions
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| status | VARCHAR(20) | pending/processing/completed/failed |
| original_url | VARCHAR(500) | MinIO path |
| svg_url | VARCHAR(500) | MinIO path |
| thumbnail_url | VARCHAR(500) | MinIO path |
| file_size_in | BIGINT | bytes |
| file_size_out | BIGINT | bytes |
| path_count | INT | SVG path elements |
| color_count | INT | unique colors |
| format_in | VARCHAR(10) | png/jpg/webp |
| error_message | TEXT | failure reason |
| created_at | TIMESTAMPTZ | DEFAULT now() |
| completed_at | TIMESTAMPTZ | |

INDEX (user_id, status), INDEX (created_at DESC)

### daily_quotas
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| date | DATE | DEFAULT CURRENT_DATE |
| count | INT | DEFAULT 0 |

UNIQUE(user_id, date)

## API Endpoints

### Auth
- `GET /api/v1/auth/github/login` — redirect to GitHub OAuth
- `GET /api/v1/auth/github/callback` — handle callback, set JWT cookie
- `GET /api/v1/auth/google/login` — redirect to Google OAuth
- `GET /api/v1/auth/google/callback` — handle callback, set JWT cookie
- `POST /api/v1/auth/refresh` — refresh JWT
- `GET /api/v1/auth/me` — current user info

### Conversions (JWT required)
- `POST /api/v1/conversions` — upload image, start conversion
- `GET /api/v1/conversions` — list user conversions (?page=&limit=&status=)
- `GET /api/v1/conversions/:id` — get single conversion status/metadata
- `GET /api/v1/conversions/:id/download` — download SVG file
- `DELETE /api/v1/conversions/:id` — soft delete

### Quota (JWT required)
- `GET /api/v1/quotas/daily` — today's usage and limit

### Middleware Chain
Request → Logger → CORS → RateLimit (100/min) → JWT Auth → Handler

## Frontend Routes

| Route | Page | Auth |
|-------|------|------|
| `/` | LandingPage (hero, features, CTA) | No |
| `/callback` | OAuthCallback (handle redirect) | No |
| `/workspace/convert` | ConvertPage (dropzone, upload, processing) | Yes |
| `/workspace/preview/:id` | PreviewPage (comparison, download) | Yes |
| `/workspace/library` | LibraryPage (history grid, filters) | Yes |

### Key Components
- **DropZone** — drag & drop + click upload, file type/size validation
- **ComparisonView** — side-by-side original vs SVG rendering
- **ZoomControls** — + / - / 100% / fit-to-screen
- **MetadataCard** — path count, color count, size reduction %
- **ConversionCard** — thumbnail, status badge, date, click to preview

## Project Structure

```
vibe-coding-svg/
├── server/
│   ├── cmd/api/main.go
│   ├── cmd/worker/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/        # auth.go, conversion.go, quota.go
│   │   ├── middleware/      # jwt, cors, ratelimit, logger
│   │   ├── model/           # user, conversion, quota
│   │   ├── repo/            # PostgreSQL data access
│   │   ├── service/         # business logic, storage (MinIO)
│   │   ├── worker/          # queue.go, vectorize.go
│   │   └── router/
│   └── migrations/
├── web/
│   ├── src/
│   │   ├── api/             # axios wrappers
│   │   ├── components/      # shared components
│   │   ├── pages/           # page components
│   │   ├── hooks/           # useAuth, useConversion, usePolling
│   │   ├── context/         # AuthContext
│   │   └── App.tsx
│   ├── vite.config.ts
│   └── tailwind.config.js
├── docker-compose.yml       # postgres + redis + minio
├── Dockerfile.api
└── Dockerfile.worker
```

## Error Handling

- Upload failures: file too large (>10MB), invalid format → 400 with user-facing message
- Conversion failures: engine crash, timeout → status=failed, error_message set, user can retry
- Auth failures: expired token → 401, redirect to login
- Rate limit: 429 with Retry-After header
- All errors return `{error: {code, message}}` JSON

## Testing Strategy

- Go: unit tests for service/repo layers, integration tests for API handlers
- React: component tests (Vitest), API mock (MSW)
- E2E: optional, can add Playwright for critical path (upload → preview → download)
