## 1. Infrastructure & Dependencies

- [ ] 1.1 docker-compose 新增 Neo4j 容器 (neo4j:5-community, ports 7474/7687, volume neo4jdata, env NEO4J_AUTH)
- [ ] 1.2 Go 添加 neo4j-go-driver/v5 依赖，创建 Neo4j driver 初始化模块 (连接池 max 20)
- [ ] 1.3 Go 添加 SVG 颜色提取工具函数 (encoding/xml 解析 fill/stroke，过滤 none/currentColor/url(#...)/transparent)
- [ ] 1.4 .env/.env.example 添加 NEO4J_PASSWORD 配置项

## 2. PostgreSQL Data Layer

- [ ] 2.1 创建 migration 007: icons 表 (id, user_id FK, name, svg_content, is_public DEFAULT false, download_count DEFAULT 0, created_at, updated_at)
- [ ] 2.2 创建 migration 008: tags 字典表 (id, name, slug UNIQUE, type CHECK usage/style/category, usage_count) + icon_tags + icon_colors + icon_themes 关联表
- [ ] 2.3 创建 model: Icon, Tag, IconTag, IconColor, IconTheme 结构体 (Tag 含 Type 字段，Icon 含 IsPublic 字段)
- [ ] 2.4 创建 repo: IconRepo (Create, BatchCreate, FindByID, FindByUserID, Delete, SetVisibility)
- [ ] 2.5 创建 repo: TagRepo (FindOrCreate by slug 去重, List, IncrementUsageCount)
- [ ] 2.6 创建 repo: SearchIcons (PG ts_vector on name + tag filter AND + color HSL ±15 + theme + is_public=true + sort + pagination)

## 3. Neo4j Graph Layer

- [ ] 3.1 创建 Neo4j session/transaction 封装，driver 单例 + 连接池管理
- [ ] 3.2 创建 graph repo: CreateIconNode + CreateIconRelations (MERGE Tag/Color/Theme 节点，Tag 含 type 属性，CREATE HAS_TAG/HAS_COLOR/IN_THEME 边)
- [ ] 3.3 创建 graph repo: DeleteIconNode + 清理孤立关系
- [ ] 3.4 创建 graph repo: GetRelatedIcons (Cypher 含 type 加权: usage×3, style×2, category×1, color×1, theme×1，仅返回公开图标)
- [ ] 3.5 创建 graph sync service: 封装双写逻辑，异步 goroutine + 3 次重试 + 失败日志

## 4. API Layer

- [ ] 4.1 创建 handler: POST /icons (Create) + POST /icons/batch (BatchCreate, max 50) — 参数校验 + 颜色自动提取
- [ ] 4.2 创建 handler: GET /icons (List, 仅公开) + GET /icons/:id (Get, 公开或所有者) + DELETE /icons/:id (Delete, 仅所有者)
- [ ] 4.3 创建 handler: GET /icons/search (关键词 + 标签 + 颜色 + 主题 + 排序 + 分页，仅返回 is_public=true)
- [ ] 4.4 创建 handler: GET /icons/:id/recommend (Neo4j 加权推荐 → PG 补全 SVG 内容和元信息，仅公开图标)
- [ ] 4.5 创建 handler: GET /tags (返回标签字典，按 usage_count 排序，用于标签云和输入建议)
- [ ] 4.6 创建 service 层: IconService 编排 IconRepo + TagRepo + GraphSyncService
- [ ] 4.7 router 注册 /api/v1/icons 和 /api/v1/tags 路由组，icons 搜索/详情/推荐为公开接口其余需 JWT

## 5. Seed Data

- [ ] 5.1 生成约 80 个基础 SVG 图标 (line ~50, filled ~30, 覆盖 20+ usage 标签, 统一 5 色色板)
- [ ] 5.2 编写 seed 脚本: 读取 SVG 文件 → 批量调用 POST /icons/batch 灌入 PG 和 Neo4j
- [ ] 5.3 docker-compose 启动后自动跑 seed（或提供手动执行命令）

## 6. Frontend — API Client

- [ ] 6.1 API client 新增 icons 模块 (create, batchCreate, list, get, search, recommend, delete) 和 tags 模块 (list)
- [ ] 6.2 API client 类型定义: Icon, Tag, IconSearchParams, RecommendResult 等 TypeScript 类型

## 7. Frontend — Icon Library Page

- [ ] 7.1 创建 IconCard 组件: aspect-video SVG 内联缩略图 + 名称 + 首标签，点击跳详情
- [ ] 7.2 创建 SearchBar 组件: 关键词输入框 + 热门标签云 (从 GET /tags 加载)
- [ ] 7.3 创建 IconLibraryPage: SearchBar + 图标网格 (3-4 col) + 加载更多分页，URL 参数同步
- [ ] 7.4 App.tsx 添加路由 `/workspace/icons`

## 8. Frontend — Icon Detail Page

- [ ] 8.1 创建 IconDetailPage: 获取图标详情 + 调用 recommend API
- [ ] 8.2 左侧: 大尺寸 SVG 预览 (带背景网格)
- [ ] 8.3 右侧: 名称 + 标签 (按 type 分组显示) + 颜色色块 + 主题 + 下载次数
- [ ] 8.4 下方: 「相关图标」标题 + 推荐图标网格 (IconCard 复用)，空状态处理
- [ ] 8.5 App.tsx 添加路由 `/workspace/icons/:id`

## 9. Editor Integration

- [ ] 9.1 Navbar 组件添加「图标库」导航链接
- [ ] 9.2 EditorToolbar 添加「发布到图标库」按钮
- [ ] 9.3 创建 PublishDialog 组件: 名称预填 + 标签输入 (建议列表 + 自由输入) + 主题选择 + 公开/私有勾选框
- [ ] 9.4 EditorPage 集成 PublishDialog → 调用 POST /icons → toast + 跳转详情链接

## 10. Verification

- [ ] 10.1 docker-compose up 全栈启动，验证 Neo4j 容器正常 (http://localhost:7474)
- [ ] 10.2 种子数据导入 → 验证 PG icons 表 + Neo4j 节点关系数量一致
- [ ] 10.3 搜索: q=home → 返回关联图标 → 点击详情 → 下方显示推荐列表
- [ ] 10.4 编辑器发布图标 → 选择公开 → 图标库页面可见 → 详情页推荐中出现
- [ ] 10.5 私有图标: 非所有者搜索和详情均不可见
