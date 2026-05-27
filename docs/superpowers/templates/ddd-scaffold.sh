#!/usr/bin/env bash
# DDD Directory Scaffold — create the four-layer structure for a new bounded context
# Usage: bash scripts/ddd-scaffold.sh <bounded-context-name>
# Example: bash scripts/ddd-scaffold.sh conversion

set -euo pipefail

BC="${1:-}"
if [ -z "$BC" ]; then
    echo "Usage: $0 <bounded-context-name>"
    exit 1
fi

echo "Scaffolding DDD structure for bounded context: $BC"

# Backend layers
mkdir -p "internal/domain/$BC"
mkdir -p "internal/service/$BC"
mkdir -p "internal/infrastructure/persistence"
mkdir -p "internal/infrastructure/storage"
mkdir -p "internal/interfaces/http"
mkdir -p "internal/interfaces/worker"

# Frontend feature directory
mkdir -p "web/src/features/$BC"
mkdir -p "web/src/domain"

# Docs placeholder
mkdir -p "docs/speckit/$BC"

# Domain layer files
cat > "internal/domain/$BC/entity.go" <<GOEOF
package $BC

// Entity — domain object with an identity (ID).
// Replace with actual fields.
type Entity struct {
    ID string
}
GOEOF

cat > "internal/domain/$BC/value_object.go" <<GOEOF
package $BC

// ValueObject — immutable, no identity. Equality by value.
// Replace with actual fields.
type ValueObject struct {
    Value string
}
GOEOF

cat > "internal/domain/$BC/aggregate.go" <<GOEOF
package $BC

// Aggregate — transactional boundary. All mutations go through the root.
// Replace with actual fields and business logic.
type Aggregate struct {
    root Entity
}
GOEOF

cat > "internal/domain/$BC/repository.go" <<GOEOF
package $BC

import "context"

// Repository interface — defined in domain, implemented in infrastructure.
type Repository interface {
    Save(ctx context.Context, aggregate *Aggregate) error
    FindByID(ctx context.Context, id string) (*Aggregate, error)
}
GOEOF

# Service layer file
cat > "internal/service/$BC/command.go" <<GOEOF
package $BC

import (
    "context"
    "your-project/internal/domain/$BC"
)

// Command is the input DTO for a use case.
type Command struct {
    // fill: input fields
}

// Handler orchestrates the use case.
type Handler struct {
    repo $BC.Repository
}

func NewHandler(repo $BC.Repository) *Handler {
    return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
    // fill: orchestrate domain logic
    return nil
}
GOEOF

echo ""
echo "Scaffold complete. Next steps:"
echo "  1. Run TaskMaster to generate docs/speckit/$BC/tasks.md"
echo "  2. Quinn writes tests → Pirlo implements → QA gate"
echo "  3. Remember: domain/ NEVER imports external libraries (§1)"
