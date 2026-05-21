# vibe-coding-svg

Bitmap to SVG conversion service. Upload a PNG/JPEG, get an SVG back.

## Architecture

```
Browser ──▶ API (:8080) ──▶ PostgreSQL    (users, conversions, quotas)
                │                   
                │  enqueue task     
                ▼                   
             Redis ──────▶ Worker ──▶ vtracer CLI
                                 │
                                 ▼
                              MinIO    (originals + SVG results)
```

Five services: **PostgreSQL** (data), **Redis** (queue), **MinIO** (files), **API** (HTTP), **Worker** (conversion). API and Worker are independent processes that communicate via Redis.

## Quick Start

```bash
git clone https://github.com/fan1ai2/vibe-coding-svg.git
cd vibe-coding-svg

# Set OAuth credentials (optional, skip for dev)
export GITHUB_CLIENT_ID=xxx
export GITHUB_CLIENT_SECRET=xxx
export GOOGLE_CLIENT_ID=xxx
export GOOGLE_CLIENT_SECRET=xxx

docker-compose up -d
```

That's it. Migrations run automatically on startup. Visit `http://localhost:8080/health` to verify.

## API

### Auth

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/auth/github/login` | - | Redirect to GitHub OAuth |
| GET | `/api/v1/auth/github/callback` | - | GitHub OAuth callback |
| GET | `/api/v1/auth/google/login` | - | Redirect to Google OAuth |
| GET | `/api/v1/auth/google/callback` | - | Google OAuth callback |
| POST | `/api/v1/auth/refresh` | JWT | Refresh token |
| GET | `/api/v1/auth/me` | JWT | Get current user info |

### Conversions

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/conversions` | JWT | Upload image (multipart) |
| GET | `/api/v1/conversions` | JWT | List conversions |
| GET | `/api/v1/conversions/:id` | JWT | Get conversion status |
| GET | `/api/v1/conversions/:id/download` | JWT | Download SVG result |

### Health

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | DB ping check |

### Usage

```bash
# Login via browser
open http://localhost:8080/api/v1/auth/github/login

# After login, use the JWT token from the callback
TOKEN="eyJ..."

# Upload
curl -X POST http://localhost:8080/api/v1/conversions \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@photo.png"

# Check status
curl http://localhost:8080/api/v1/conversions/<id> \
  -H "Authorization: Bearer $TOKEN"

# Download
curl http://localhost:8080/api/v1/conversions/<id>/download \
  -H "Authorization: Bearer $TOKEN" \
  -o result.svg
```

Conversion statuses: `pending` → `processing` → `completed` (or `failed`).

## Configuration

All via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | API server port |
| `DATABASE_URL` | `postgres://svguser:svgpass@localhost:5432/svgconverter?sslmode=disable` | PostgreSQL connection |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO endpoint |
| `MINIO_ACCESS_KEY` | `minioadmin` | MinIO access key |
| `MINIO_SECRET_KEY` | `minioadmin` | MinIO secret key |
| `JWT_SECRET` | `dev-secret-change-in-prod` | JWT signing secret |
| `GITHUB_CLIENT_ID` | - | GitHub OAuth app client ID |
| `GITHUB_CLIENT_SECRET` | - | GitHub OAuth app secret |
| `GOOGLE_CLIENT_ID` | - | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | - | Google OAuth client secret |

## Development

```bash
# Start infrastructure
docker-compose up -d postgres redis minio

# Run API
cd server
go run ./cmd/api/

# Run Worker (another terminal)
go run ./cmd/worker/
```

Requires **vtracer** installed on `$PATH` for the worker:

```bash
# macOS
brew install vtracer

# Linux
cargo install vtracer
```

## Project Structure

```
server/
├── cmd/
│   ├── api/main.go              # API entry point
│   └── worker/main.go           # Worker entry point
├── internal/
│   ├── config/config.go         # Env-based config
│   ├── model/                   # Data types
│   │   ├── user.go
│   │   ├── conversion.go
│   │   └── quota.go
│   ├── repo/                    # Database layer
│   │   ├── user.go
│   │   └── conversion.go
│   ├── service/                 # Business logic
│   │   ├── auth.go              # OAuth + JWT
│   │   ├── storage.go           # MinIO wrapper
│   │   └── conversion.go        # Conversion pipeline
│   ├── handler/                 # HTTP handlers
│   │   ├── auth.go
│   │   ├── conversion.go
│   │   └── health.go
│   ├── middleware/              # Gin middleware
│   │   ├── jwt.go
│   │   ├── cors.go
│   │   └── ratelimit.go
│   ├── router/router.go        # Route wiring
│   ├── worker/                 # Background worker
│   │   ├── converter.go        # vtracer wrapper
│   │   └── worker.go           # asynq handler
│   └── migrate/migrate.go      # Auto-migration
├── migrations/                  # SQL migration files
├── Dockerfile.api
├── Dockerfile.worker
├── go.mod
└── go.sum
```
