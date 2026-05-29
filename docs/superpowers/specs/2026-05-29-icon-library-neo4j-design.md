# Icon Library with Neo4j Graph Search — Design Spec

## Overview

在 SVG 资源工坊中新增图标库管理系统。PostgreSQL 存图标元数据和 SVG 内容，Neo4j 管理图标间基于标签、颜色、主题的网状关系图，支持多维度检索和关联推荐。

## Architecture

```
PostgreSQL（业务数据）               Neo4j（图关系引擎）
┌──────────────────────┐            ┌──────────────────────────────┐
│ icons                │  异步双写   │ (:Icon {id, name})            │
│ tags (分层字典)       │ ←────────→ │   ├─ [:HAS_TAG] → (:Tag)     │
│ icon_tags            │            │   ├─ [:HAS_COLOR] → (:Color)  │
│ icon_colors          │            │   ├─ [:IN_THEME] → (:Theme)   │
│ icon_themes          │            │   └─ [:RELATED_TO] → (:Icon)  │
│ is_public 控制可见性   │            └──────────────────────────────┘
└──────────────────────┘
```

**同步策略**: 写入 API 在 PG 事务提交后即返回 201，Neo4j 操作异步 goroutine 执行，失败重试 3 次，不一致通过补偿任务修复。

**与 saved_svgs 关系**: icons 是独立表，saved_svgs 保持为私有草稿。编辑器「发布到图标库」从 saved_svg 或当前编辑内容创建 icon 记录。

## Tag System

三级分层标签，在 Neo4j 推荐中权重不同：

| 类型 | 说明 | 示例 | 推荐权重 |
|------|------|------|----------|
| `usage` | 用途语义 | home, arrow, chart, user, lock | ×3 |
| `style` | 视觉风格 | flat, line, filled, gradient, duotone | ×2 |
| `category` | 图标分类 | UI, logo, illustration, brand | ×1 |

标签管理：预设基础标签字典随种子数据入库。用户可扩展新标签，自由输入但写入前做同名去重。

## Data Tables (PostgreSQL)

```sql
-- 图标主表
icons (id UUID PK, user_id UUID FK, name VARCHAR, svg_content TEXT,
       is_public BOOL DEFAULT false, download_count INT DEFAULT 0,
       created_at, updated_at)

-- 标签字典
tags (id UUID PK, name VARCHAR, slug VARCHAR UNIQUE,
      type VARCHAR CHECK(type IN ('usage','style','category')))

-- 图标-标签 多对多
icon_tags (icon_id UUID FK, tag_id UUID FK, PRIMARY KEY(icon_id, tag_id))

-- 图标主色（自动提取）
icon_colors (icon_id UUID FK, color_hex VARCHAR, role VARCHAR, PRIMARY KEY(icon_id, color_hex))

-- 主题关联
icon_themes (icon_id UUID FK, theme_name VARCHAR, PRIMARY KEY(icon_id, theme_name))
```

## Neo4j Graph Model

```
Nodes:  (:Icon), (:Tag), (:Color), (:Theme)
Edges:  [:HAS_TAG]    {type}        Icon → Tag
        [:HAS_COLOR]  {role}        Icon → Color
        [:IN_THEME]                 Icon → Theme
        [:RELATED_TO] {reason}      Icon → Icon
```

Isolated nodes (no remaining relations) are retained after icon deletion — they may be reused by future icons.

## Recommendation Engine

Cypher query on icon detail page load:

```cypher
MATCH (i:Icon {id: $id})-[r:HAS_TAG|:HAS_COLOR|:IN_THEME]->(attr)
     <-[:HAS_TAG|:HAS_COLOR|:IN_THEME]-(related:Icon)
WHERE related.id <> i.id
WITH related, attr, r,
     CASE type(r)
       WHEN 'HAS_TAG' THEN
         CASE attr.type WHEN 'usage' THEN 3 WHEN 'style' THEN 2 ELSE 1 END
       ELSE 1
     END AS weight
RETURN related.id, related.name, sum(weight) AS score
ORDER BY score DESC LIMIT 10
```

推荐结果从 PG 补全 SVG 内容再返回前端。初版不做 TF-IDF 去热门标签偏差，先上线看效果。

## API Design

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/icons` | JWT | 创建单个图标 |
| POST | `/api/v1/icons/batch` | JWT | 批量创建 (max 50) |
| GET | `/api/v1/icons` | JWT | 分页查询图标列表 |
| GET | `/api/v1/icons/search` | - | 搜索: q + tags + color + theme + sort |
| GET | `/api/v1/icons/:id` | - | 图标详情 |
| GET | `/api/v1/icons/:id/recommend` | - | 关联推荐 (Neo4j) |
| DELETE | `/api/v1/icons/:id` | JWT | 删除图标 (仅所有者) |

公开接口（search/detail/recommend）需公开图标可见性控制：仅返回 `is_public=true` 的图标。

## Pages

### Icon Library Page (`/workspace/icons`)

```
┌─────────────────────────────────────────┐
│  🔍 搜索图标...                           │
│  [home] [flat] [line] [arrow] [user]...  │  ← 热门标签云
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  │
│  │ SVG  │ │ SVG  │ │ SVG  │ │ SVG  │  │  ← 图标网格 (3-4 col)
│  │ name  │ │ name  │ │ name  │ │ name  │  │
│  └──────┘ └──────┘ └──────┘ └──────┘  │
│  加载更多                                 │
└─────────────────────────────────────────┘
```

空搜索框 + 热门标签云 + 图标网格。点击图标卡片进入详情页。

### Icon Detail Page (`/workspace/icons/:id`)

```
┌─────────────────────────────────────────┐
│  ┌─────────┐  名称: house.svg            │
│  │         │  标签: [home] [flat] [UI]   │
│  │  SVG    │  颜色: #FF0000 #00FF00      │
│  │  大图   │  下载: 42 次                 │
│  └─────────┘                             │
│                                          │
│  相关图标                                 │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  │
│  │ SVG  │ │ SVG  │ │ SVG  │ │ SVG  │  │
│  └──────┘ └──────┘ └──────┘ └──────┘  │
└─────────────────────────────────────────┘
```

上方大图 + 元信息，下方关联推荐网格。推荐项可点击跳转到对应详情页。

## Seed Data

约 80 个基础 SVG 图标，覆盖 UI 常用场景：

- **Usage 标签**: home, user, search, settings, mail, lock, cart, heart, star, share, download, upload, edit, delete, add, close, check, arrow, menu, filter, bell, bookmark, camera, clock, calendar, folder, file, map, phone, message
- **Style 标签**: line (50 个), filled (30 个)
- **Category**: 全部归入 UI
- **Colors**: 种子图标使用统一色板 (#3B82F6, #6B7280, #10B981, #EF4444, #F59E0B)

种子图标由 Claude 生成 SVG 内容，通过批量 API 灌入 PG 和 Neo4j。

## Color Extraction

Go 端 `encoding/xml` 解析 SVG，遍历所有元素提取 `fill` / `stroke` 属性值（去除非颜色值如 `none`、`currentColor`、`url(#...)`），按出现频率排序取前 5 个颜色写入 `icon_colors`。

## Infrastructure

docker-compose.yml 新增 Neo4j 服务：

```yaml
neo4j:
  image: neo4j:5-community
  ports:
    - "7474:7474"
    - "7687:7687"
  environment:
    NEO4J_AUTH: neo4j/${NEO4J_PASSWORD}
  volumes:
    - neo4jdata:/data
```

Go 新依赖: `github.com/neo4j/neo4j-go-driver/v5`

## Implementation Order

1. Infrastructure — Neo4j 容器 + Go driver + 颜色提取工具
2. PG Data Layer — migrations + models + repos
3. Neo4j Graph Layer — 节点/关系 CR + Cypher 查询封装 + 异步双写服务
4. API Layer — icon CRUD + search + recommend handlers + router
5. Frontend — IconLibraryPage + IconDetailPage + IconCard + SearchBar + Navbar 入口
6. Editor Integration — 发布按钮 + 弹窗 → 创建 API
7. Seed Data — 生成 80 个 SVG + 批量导入脚本

---

**Design approved 2026-05-29. Ready for implementation plan.**
