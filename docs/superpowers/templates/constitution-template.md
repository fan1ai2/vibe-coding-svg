# [Project Name] Constitution

## §1 Architecture Principles

- DDD tactical layering: `domain/` → `service/` → `infrastructure/` → `interfaces/`
- Dependency direction: `interfaces → service → domain ← infrastructure`
- `domain/` MUST NOT import any external library (database drivers, Redis, S3, HTTP frameworks) — standard library and self-defined interfaces only
- `service/` depends ONLY on `domain/`
- Each bounded context is an independent package
- Dependency injection happens exclusively in `cmd/server/main.go`

## §2 Quality Gate Standards

- Build: zero errors (`go build ./...` + `tsc -b`)
- Test: all tests pass (`go test ./...` + frontend test runner)
- Coverage: domain layer test coverage ≥ 80%
- Integration: at least 1 happy-path test per API endpoint
- All Speckit checklist items checked before merge

## §3 Technology Stack

<!-- Fill in with actual project stack. Example: -->
<!-- Backend: Go 1.25 + Gin -->
<!-- Frontend: React 19 + TypeScript 5 + Vite -->
<!-- Database: PostgreSQL 16 -->
<!-- Cache/Queue: Redis 7 -->
<!-- Storage: MinIO (S3-compatible) -->
<!-- Container: Docker Compose -->

| Layer | Technology | Constraint |
|-------|-----------|------------|
| Backend | <!-- fill --> | <!-- fill --> |
| Frontend | <!-- fill --> | <!-- fill --> |
| Database | <!-- fill --> | <!-- fill --> |
| Cache/Queue | <!-- fill --> | <!-- fill --> |
| Storage | <!-- fill --> | <!-- fill --> |

## §4 Workflow Rules

- Every feature MUST have a written spec (`docs/superpowers/specs/*.spec.md`) before any code
- Spec → TaskMaster task breakdown → Speckit clarify → Speckit plan → TDD implement → QA gate
- No phase may be skipped
- All phase artifacts MUST be git committed
- Implementation follows TDD: QA writes tests first (red), Dev writes implementation (green)
- Domain layer is built first; infrastructure layer is built last

## §5 Prohibitions

- No code before spec
- No external imports in `domain/` layer
- No cross-bounded-context PRs (one BC per PR)
- No "incidental refactoring" without a spec
- No YAGNI abstractions — three similar lines is fine until a pattern proves itself

**Version**: 1.0.0 | **Ratified**: <!-- fill date --> | **Last Amended**: <!-- fill date -->
