## Context

SVG 资源工坊已有位图转 SVG 能力和颜色编辑器，但缺少图标资产的持久化存储和智能检索。图标体积小（通常几 KB），天然适合批量管理和图形化检索。引入 Neo4j 作为图数据库，实现图标间基于标签、颜色、主题的多维度关联推荐。

当前技术栈：Go 1.25 + Gin + PostgreSQL 16 + Redis + MinIO + React 19 + TypeScript 5。Neo4j 和 neo4j-go-driver 是新依赖。

## Goals / Non-Goals

**Goals:**
- PostgreSQL 存储图标元数据和 SVG 内容，支持批量写入
- Neo4j 构建图标节点及其标签、颜色、主题关系图，支持加权关联推荐
- 多维度检索：关键词 + 标签组合 + 颜色色相范围 + 主题筛选
- 访问控制：图标默认私有，发布时可设为公开，公开图标对所有人可见
- 前端图标库页面：搜索 + 热门标签云 + 图标网格；详情页：大图预览 + 关联推荐
- 编辑器到图标库的发布链路
- 种子数据：约 80 个基础 UI SVG 图标，覆盖 line/filled 风格

**Non-Goals:**
- 不引入 Elasticsearch —— PostgreSQL ts_vector + ILIKE 够用
- 不做图片相似度搜索 —— 纯属性图关联，不涉及向量
- 不替换现有 saved_svgs 表 —— 图标库是独立概念，saved_svgs 保持为私有草稿
- 不做 TF-IDF 热门标签去偏 —— 先上线看效果，后期再调

## Decisions

### 1. PostgreSQL + Neo4j 双库模式

PG 是主数据源（用户、图标元数据、SVG 内容），Neo4j 是关系引擎。写入时双写：PG 写完后，异步写 Neo4j 节点和关系边。Neo4j 写入失败不影响 PG 事务结果，通过后台补偿保证最终一致。

**替代方案**: 纯 PG 关联表 —— 能跑但多跳遍历 + 共享属性计分排序会变成复杂子查询，不如 Cypher 一句干净。

### 2. Neo4j 节点模型

```
(:Icon {id, name})
  -[:HAS_TAG]-> (:Tag {name, slug, type})
  -[:HAS_COLOR]-> (:Color {hex, role})
  -[:IN_THEME]-> (:Theme {name})
  -[:RELATED_TO {reason}]-> (:Icon)
```

`type` 字段区分标签维度，在推荐查询中赋予不同权重：

| type | 示例 | 推荐权重 |
|------|------|----------|
| `usage` | home, arrow, user, lock | ×3 |
| `style` | flat, line, filled, gradient | ×2 |
| `category` | UI, logo, illustration, brand | ×1 |

**替代方案**: 扁平标签等权 —— 热门风格标签会主导推荐结果，分层权重让用途语义匹配更突出。

### 3. 标签管理

预设标签字典 + 用户扩展。种子数据导入时建立基础标签集（约 30 个）。用户创建图标时可自由输入新标签，写入前同名去重。所有标签归入 usage/style/category 三类之一，默认新标签归入 usage。

### 4. 写入流程

```
API 收到创建请求
  → PG: INSERT INTO icons + icon_tags + icon_colors + icon_themes (事务)
  → 返回 201 给客户端
  → 异步 goroutine: Neo4j MERGE 节点 + CREATE 关系
  → 失败重试 3 次，仍失败记录日志，后续通过 reconcile job 补偿
```

**替代方案**: 双写事务 —— Neo4j 不支持跨库 2PC，强一致不现实。异步最终一致更务实。

### 5. 搜索架构

```
GET /api/v1/icons/search?q=house&tags=home,building&color=red&theme=business&sort=popular
```

- 关键词匹配：PG `ts_vector` on `name` + tag names
- 标签筛选：`icon_tags` JOIN，多个标签取交集（AND 语义）
- 颜色筛选：`icon_colors` 按 HSL 色相 ±15 度范围匹配
- 主题筛选：`icon_themes` 直接匹配
- 排序：`download_count DESC` / `created_at DESC` / 匹配相关度
- 可见性：仅返回 `is_public=true` 的图标（公开接口）；已登录用户可查看自己的私有图标
- 分页：limit + offset

搜索页入口：空搜索框 + 热门标签云（按 usage_count 排序）+ 图标网格。点击卡片进入详情页。

### 6. 关联推荐

```
GET /api/v1/icons/:id/recommend?limit=10
```

Cypher 查询（含标签权重）：
```cypher
MATCH (i:Icon {id: $id})-[r:HAS_TAG|:HAS_COLOR|:IN_THEME]->(attr)
     <-[:HAS_TAG|:HAS_COLOR|:IN_THEME]-(related:Icon)
WHERE related.id <> i.id
WITH related, r, attr,
     CASE type(r)
       WHEN 'HAS_TAG' THEN
         CASE attr.type WHEN 'usage' THEN 3 WHEN 'style' THEN 2 ELSE 1 END
       ELSE 1
     END AS weight
RETURN related.id, related.name, sum(weight) AS score
ORDER BY score DESC LIMIT $limit
```

返回时从 PG 补全 SVG 内容。推荐结果仅包含公开图标。

### 7. 批量入库与颜色提取

Go 端用 `encoding/xml` 解析 SVG，遍历所有元素提取 `fill` / `stroke` 属性值。过滤掉：`none`、`currentColor`、`url(#...)`、`transparent`。按出现频率排序取前 5 个写入 `icon_colors`，同步到 Neo4j `HAS_COLOR` 关系。

### 8. 访问控制

- `icons.is_public` 字段：默认 `false`
- 编辑器「发布到图标库」时弹窗提供「公开到图标库」勾选项
- 公开列表/搜索/推荐/详情接口仅返回 `is_public=true` 的图标
- 已登录用户可见自己的私有图标（通过 JWT user_id 过滤）
- 删除操作限制图标所有者

### 9. 种子数据

约 80 个基础 SVG 图标，由 Claude 生成 SVG 内容：

- **Usage 标签** (20+): home, user, search, settings, mail, lock, cart, heart, star, share, download, upload, edit, delete, add, close, check, arrow, menu, filter, bell, bookmark, camera, clock, calendar, folder, file, map, phone, message
- **Style**: line (~50), filled (~30)
- **Category**: 全部 UI
- **Colors**: 统一色板 (#3B82F6, #6B7280, #10B981, #EF4444, #F59E0B)
- 通过 `POST /api/v1/icons/batch` 一次性灌入

## Risks / Trade-offs

- [Neo4j 写入失败导致数据不一致] → 异步补偿 + reconcile cron job 每小时校验
- [Neo4j 社区版单机无高可用] → 图标量级初期可控，需要时升企业版
- [双写增加延迟] → 写入 API 在 PG 事务提交后即返回，Neo4j 异步不收影响
- [热门标签主导推荐] → 分层权重缓解，不上 TF-IDF，先上线看效果再迭代
- [种子图标质量] → Claude 生成简单几何图形 SVG，用户后期可替换或添加
