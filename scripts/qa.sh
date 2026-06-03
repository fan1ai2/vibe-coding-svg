#!/usr/bin/env bash
# SVG 资源工坊 — 一键质量检测
# 用法: ./scripts/qa.sh
set -euo pipefail
exec bash "$(dirname "$0")/../.specify/scripts/bash/quality-gate.sh"
