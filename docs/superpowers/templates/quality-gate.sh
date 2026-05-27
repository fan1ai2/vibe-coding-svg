#!/usr/bin/env bash
# Quality Gate — one command to validate the full stack
# Usage: bash scripts/qa.sh
#
# Adapt the build/test commands below to match your project's tech stack.
# This template assumes: Go backend + TypeScript/Node frontend.
# Remove or comment out sections that don't apply.

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

PASS=0
FAIL=0

check() {
    local label="$1"
    shift
    echo -n "  $label ... "
    if "$@" > /dev/null 2>&1; then
        echo -e "${GREEN}PASS${NC}"
        ((PASS++))
    else
        echo -e "${RED}FAIL${NC}"
        ((FAIL++))
    fi
}

echo "=== Quality Gate ==="
echo ""

# ---- Spec Compliance ----
check "openspec validate" openspec validate

# ---- Backend Build ----
check "go build ./..." go build ./...

# ---- Backend Tests ----
check "go test ./..." go test ./...

# ---- Backend Coverage (domain layer only) ----
echo -n "  domain coverage ≥ 80% ... "
COVER=$(go test ./internal/domain/... -cover 2>/dev/null | grep -oP 'coverage: \K[0-9.]+' | head -1 || echo "0")
if (( $(echo "$COVER >= 80" | bc -l 2>/dev/null || echo 0) )); then
    echo -e "${GREEN}PASS (${COVER}%)${NC}"
    ((PASS++))
else
    echo -e "${RED}FAIL (${COVER}% < 80%)${NC}"
    ((FAIL++))
fi

# ---- Frontend Type Check ----
check "tsc -b" npx tsc -b

# ---- Frontend Build ----
check "npm run build" npm run build

# ---- Frontend Tests ----
check "frontend tests" npx vitest --run

echo ""
echo "=== Result: $PASS passed, $FAIL failed ==="

if [ "$FAIL" -gt 0 ]; then
    echo ""
    echo "Failure paths:"
    echo "  Build failure   → Pirlo fixes code, re-run QA"
    echo "  Test failure    → Pirlo fixes implementation, re-run QA"
    echo "  Coverage < 80%  → Quinn adds domain tests, re-run QA"
    echo "  Spec mismatch   → Return to Speckit plan phase, realign"
    exit 1
fi

echo "All gates passed."
