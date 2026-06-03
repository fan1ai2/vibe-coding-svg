#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# SVG 资源工坊 — Quality Gate Script
# 整合 openspec validate + BMAD QA + type check + build
# ============================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

check() {
  local name="$1"
  shift
  echo -e "${YELLOW}[CHECK]${NC} $name ..."
  if "$@"; then
    echo -e "  ${GREEN}PASS${NC}"
    ((PASSED++))
  else
    echo -e "  ${RED}FAIL${NC}"
    ((FAILED++))
  fi
  echo
}

echo "=============================================="
echo " Quality Gate — $(date '+%Y-%m-%d %H:%M:%S')"
echo "=============================================="
echo

# ---- Gate 1: OpenSpec Validate ----
check "openspec validate (spec compliance)" openspec validate 2>&1 || true

# ---- Gate 2: Go Build (API + Worker) ----
check "Go build server/..." bash -c "cd server && go build ./cmd/api/ && go build ./cmd/worker/"

# ---- Gate 3: TypeScript Check ----
check "TypeScript compilation (web)" bash -c "cd web && npx tsc -b"

# ---- Gate 4: Frontend Build ----
check "Vite build (web)" bash -c "cd web && npm run build"

echo "=============================================="
echo -e " Results: ${GREEN}${PASSED} passed${NC}, ${RED}${FAILED} failed${NC}"
echo "=============================================="

if [ "$FAILED" -gt 0 ]; then
  exit 1
fi
