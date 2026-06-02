.PHONY: seed seed-bin build-api build-web build build-push push up down up-prod pull-prod

# ---- 本地开发 ----

# 启动全部服务（本地开发，本地编译）
up:
	docker compose up -d

# 停止全部服务
down:
	docker compose down

# 构建 API 镜像
build-api:
	docker compose build api

# 构建前端镜像
build-web:
	docker compose build web

# ---- 生产部署 ----

# 本地构建全部镜像（api + worker + web）
build:
	docker compose build api worker web

# 构建并推送全部镜像
build-push: build
	docker compose push api worker web

# 仅推送已构建的镜像
push:
	docker compose push api worker web

# [服务器] 拉取镜像
pull-prod:
	docker compose -f docker-compose.prod.yml pull

# [服务器] 启动全部服务
up-prod:
	docker compose -f docker-compose.prod.yml up -d

# [服务器] 查看状态
ps-prod:
	docker compose -f docker-compose.prod.yml ps

# [服务器] 查看日志
logs-prod:
	docker compose -f docker-compose.prod.yml logs -f

# 编译 seed 二进制
seed-bin:
	cd server && go build -o bin/seed ./cmd/seed/

# 导入种子图标数据（需先启动服务: make up）
# 自动获取 guest token 并批量导入 ~80 个基础图标
seed: seed-bin
	@echo "获取 guest token..."
	@TOKEN=$$(curl -sf -X POST http://localhost:8080/api/v1/auth/guest | grep -o '"token":"[^"]*"' | cut -d'"' -f4); \
	if [ -z "$$TOKEN" ]; then \
		echo "错误: 无法获取 token，请确认 API 服务已启动 (make up)"; \
		exit 1; \
	fi; \
	echo "导入种子图标..."; \
	server/bin/seed http://localhost:8080 "$$TOKEN"; \
	echo "种子数据导入完成"
