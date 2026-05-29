## ADDED Requirements

### Requirement: Create icon
系统 SHALL 支持创建图标，接收 SVG 内容、名称、标签、颜色、主题及公开标志，写入 PostgreSQL 并异步同步到 Neo4j。

#### Scenario: Single icon creation
- **WHEN** 已认证用户提交 name="house.svg"、svg_content、tags=["home","building"]、colors=["#FF0000","#00FF00"]、theme="business"、is_public=true
- **THEN** PostgreSQL icons 表写入一条记录（含 is_public=true），icon_tags/icon_colors/icon_themes 关联写入，返回 201 + 图标 ID

#### Scenario: Icon creation without auth
- **WHEN** 未认证用户提交创建请求
- **THEN** 系统返回 401 未授权错误

#### Scenario: SVG content too large
- **WHEN** 用户提交的 svg_content 超过 2MB
- **THEN** 系统返回 413 内容过大错误

### Requirement: Icon visibility control
系统 SHALL 支持图标公开/私有控制，公开图标对所有用户可见，私有图标仅所有者可见。

#### Scenario: Public list excludes private icons
- **WHEN** 用户 A 的图标 is_public=false
- **THEN** 用户 B 在列表和搜索中看不到该图标

#### Scenario: Owner sees own private icons
- **WHEN** 用户 A 查看自己的图标列表
- **THEN** 列表包含 is_public=false 的图标

### Requirement: Batch create icons
系统 SHALL 支持批量创建图标，一次请求创建最多 50 条记录。

#### Scenario: Batch creation
- **WHEN** 已认证用户提交包含 3 个图标的数组
- **THEN** 3 条图标记录全部写入 PG，返回 201 + 3 个图标 ID 列表

#### Scenario: Batch size limit
- **WHEN** 批量请求超过 50 个图标
- **THEN** 系统返回 400 参数错误

### Requirement: List icons
系统 SHALL 支持分页查询图标列表，返回图标元信息和 SVG 内容。

#### Scenario: List with pagination
- **WHEN** 请求 `GET /icons?limit=20&offset=0`
- **THEN** 按创建时间倒序返回最多 20 条公开图标记录，每条包含 id、name、tags、colors、theme、svg_content、is_public、download_count

#### Scenario: Empty list
- **WHEN** 数据库中无公开图标记录
- **THEN** 返回空数组 `[]`

### Requirement: Get icon by ID
系统 SHALL 支持按 ID 查询单个图标完整信息。

#### Scenario: Existing public icon
- **WHEN** 请求存在的公开图标 ID
- **THEN** 返回该图标的完整信息（元数据 + SVG 内容 + 标签 + 颜色 + 主题）

#### Scenario: Non-existing icon
- **WHEN** 请求不存在的图标 ID
- **THEN** 返回 404 记录不存在错误

#### Scenario: Private icon by non-owner
- **WHEN** 非所有者请求私有图标
- **THEN** 返回 404 记录不存在错误

### Requirement: Delete icon
系统 SHALL 支持按 ID 删除图标及其所有关联数据，仅限所有者操作。

#### Scenario: Delete own icon
- **WHEN** 图标所有者请求删除
- **THEN** PG 中图标及关联数据被删除，Neo4j 中对应节点和关系异步删除，返回 200

#### Scenario: Delete another user's icon
- **WHEN** 非所有者请求删除
- **THEN** 系统返回 403 禁止操作错误

### Requirement: Extract SVG colors
系统 SHALL 在创建图标时自动解析 SVG 内容中的 fill 和 stroke 颜色值，过滤无效颜色后提取主色。

#### Scenario: SVG with multiple fills
- **WHEN** SVG 包含 `<rect fill="#FF0000"/>` 和 `<circle fill="#00FF00"/>`
- **THEN** 系统提取 `#FF0000` 和 `#00FF00` 写入 icon_colors 表

#### Scenario: SVG with special color values
- **WHEN** SVG 包含 `fill="none"`、`stroke="currentColor"`、`fill="url(#gradient)"`、`stroke="transparent"`
- **THEN** 这些值被过滤，不写入 icon_colors

#### Scenario: SVG with no valid fill/stroke
- **WHEN** SVG 所有元素的 fill 和 stroke 均为无效值
- **THEN** icon_colors 表不写入该图标的任何颜色记录
