## ADDED Requirements

### Requirement: Full-text keyword search
系统 SHALL 支持按关键词搜索图标名称。

#### Scenario: Search by name keyword
- **WHEN** 请求 `GET /icons/search?q=house`
- **THEN** 返回名称中包含 "house" 且含标签 "home" 等的图标，按匹配度排序

#### Scenario: Search with no results
- **WHEN** 关键词无匹配
- **THEN** 返回空数组 `[]`

### Requirement: Tag filter
系统 SHALL 支持按一个或多个标签筛选图标，多标签为 AND 逻辑。

#### Scenario: Filter by multiple tags
- **WHEN** 请求 `GET /icons/search?tags=home,building`
- **THEN** 仅返回同时包含 "home" 和 "building" 标签的图标

#### Scenario: Tag filter alone
- **WHEN** 请求 `GET /icons/search?tags=flat` 不带关键词
- **THEN** 返回所有含 "flat" 标签的图标，按创建时间倒序

### Requirement: Color filter
系统 SHALL 支持按颜色筛选图标，匹配色相 ±15 度范围内的图标。

#### Scenario: Filter by color name
- **WHEN** 请求 `GET /icons/search?color=red`
- **THEN** 返回 icon_colors 中色相在 345-360 或 0-15 范围内的图标

#### Scenario: Filter by hex color
- **WHEN** 请求 `GET /icons/search?color=%23FF0000`
- **THEN** 系统解析为色相 0 度，匹配 ±15 度范围内图标

### Requirement: Theme filter
系统 SHALL 支持按主题筛选图标。

#### Scenario: Filter by theme
- **WHEN** 请求 `GET /icons/search?theme=business`
- **THEN** 仅返回主题为 "business" 的图标

### Requirement: Combined filters
系统 SHALL 支持同时使用关键词、标签、颜色、主题组合筛选。

#### Scenario: Combined search
- **WHEN** 请求 `GET /icons/search?q=arrow&tags=flat&theme=ui`
- **THEN** 返回同时满足所有筛选条件的图标

### Requirement: Sort results
系统 SHALL 支持多种排序方式。

#### Scenario: Sort by popularity
- **WHEN** 请求 `GET /icons/search?sort=popular`
- **THEN** 按 download_count 降序排列

#### Scenario: Sort by newest
- **WHEN** 请求 `GET /icons/search?sort=newest`
- **THEN** 按 created_at 降序排列

#### Scenario: Default sort
- **WHEN** 请求不指定 sort 参数
- **THEN** 默认按 created_at 降序排列
