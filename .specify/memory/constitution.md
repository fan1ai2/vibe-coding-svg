<!--
Sync Impact Report
  Version change: 0.0.0 → 1.0.0 (initial ratification)
  Added:   5 core principles, 2 constraint sections, governance
  Removed: none (first version)
  Templates requiring updates:
    ✅ .specify/templates/plan-template.md (pending verify)
    ✅ .specify/templates/spec-template.md (pending verify)
    ✅ .specify/templates/tasks-template.md (pending verify)
  Follow-up TODOs: none
-->
# SVG 资源工坊 Constitution

## Core Principles

### I. Spec-First Development (NON-NEGOTIABLE)
Every feature, bug fix, and refactor MUST begin with a written spec using `/speckit.specify` or `/opsx:propose`. No code is written before the spec is reviewed and approved. Specs define user stories, acceptance criteria, and technical boundaries. Implementation follows the spec—deviations require a spec amendment first.

### II. Full Stack Coherence
Backend (Go API + Worker), Frontend (React/TypeScript via Vite), and infrastructure (Docker Compose) MUST evolve in lockstep. Any API change that affects the frontend MUST include the corresponding frontend update. Database migrations MUST be backward-compatible and accompany the relevant code change in the same PR.

### III. Quality Gates
Every change MUST pass quality gates before merge:
- `openspec validate`—spec compliance check
- BMAD QA deep dive—review against acceptance criteria
- TypeScript compilation (`tsc -b`) and Go build (`go build ./...`)
- Existing tests pass (frontend + backend)

### IV. Security & Privacy
- All authentication tokens (JWT) use short-lived access + refresh rotation
- OAuth callbacks validate state parameter to prevent CSRF
- User file uploads are isolated in MinIO with time-limited pre-signed URLs
- SMTP verification codes use TTL enforcement and rate limiting
- CORS is restricted to known frontend origins only

### V. Simplicity & YAGNI
Start with the simplest solution that meets the spec. No preemptive abstractions, no unused generalization, no premature optimization. If three similar lines of code exist, that is fine—only refactor when a pattern proves itself across multiple use cases. Complexity must be justified in the spec.

## Technology Stack

| Layer | Technology | Constraint |
|-------|-----------|------------|
| Backend API | Go 1.25, Gin framework | Stateless HTTP, JWT middleware |
| Worker | Go 1.25, vtracer (Rust) | Redis-backed job queue |
| Frontend | React 19, TypeScript, Vite, Tailwind | SPA served by Nginx |
| Database | PostgreSQL 16 | Migrations via golang-migrate |
| Cache/Queue | Redis 7 | Task queue + optional cache |
| Storage | MinIO (S3-compatible) | Pre-signed URLs for file access |
| Proxy | Caddy 2 | Auto HTTPS via Let's Encrypt |
| Container | Docker Compose | All services in one compose file |

## Development Workflow

1. **Propose**—`/opsx:propose` or `/speckit.specify` to create a change spec
2. **Brainstorm**—DDD domain modeling via superpowers brainstorming
3. **Plan**—`/speckit.plan` to create technical implementation plan
4. **Break Down**—`task-master parse-prd` or `/speckit.tasks` to decompose into subtasks
5. **Implement**—`/speckit.implement` to execute tasks
6. **Validate**—`openspec validate` + BMAD QA review + type check + build
7. **Archive**—`/opsx:archive` to close the change

## Governance

This Constitution supersedes all other project practices. Amendments require:
1. A written proposal explaining the rationale and impact
2. Review of all dependent templates for consistency
3. Version increment per semantic versioning (MAJOR for principle removal/redefinition, MINOR for additions, PATCH for clarifications)

All PRs and code reviews MUST verify compliance with these principles.

**Version**: 1.0.0 | **Ratified**: 2026-05-27 | **Last Amended**: 2026-05-27
