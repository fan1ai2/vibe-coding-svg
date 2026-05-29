## ADDED Requirements

### Requirement: Icon library page
系统 SHALL 提供图标库浏览页面，搜索框在上方 + 热门标签云 + 图标网格，探索式入口。

#### Scenario: Page loads with public icons
- **WHEN** 用户访问 `/workspace/icons`
- **THEN** 页面顶部显示搜索框，下方显示热门标签云（按使用频率排序），再下方是公开图标网格，每张卡片渲染 SVG 缩略图 + 名称

#### Scenario: Search with keyword
- **WHEN** 用户在搜索栏输入 "house" 并按回车
- **THEN** 图标网格刷新为名称匹配 "house" 的公开图标，URL 参数更新为 `?q=house`

#### Scenario: Click tag to filter
- **WHEN** 用户点击标签云中的 "flat"
- **THEN** URL 参数追加 `tags=flat`，图标网格刷新为含 "flat" 标签的图标

#### Scenario: Pagination
- **WHEN** 图标总数超过 20 个
- **THEN** 页面底部显示「加载更多」按钮，点击后追加下一页

### Requirement: Icon card component
系统 SHALL 为每个图标渲染卡片，显示 SVG 缩略图、名称和标签。

#### Scenario: Card renders SVG thumbnail
- **WHEN** 图标数据包含 svg_content
- **THEN** 卡片在 aspect-video 容器中渲染该 SVG 内联

#### Scenario: Click card navigates to detail
- **WHEN** 用户点击图标卡片
- **THEN** 导航到 `/workspace/icons/:id` 详情页

### Requirement: Icon detail page
系统 SHALL 提供图标详情页，左侧大图预览 + 元信息，下方显示基于 Neo4j 的关联推荐图标列表。

#### Scenario: Detail page with recommend section
- **WHEN** 用户访问 `/workspace/icons/:id`
- **THEN** 页面显示大尺寸 SVG 预览、名称、标签（按类型分组）、颜色色块、下载次数、主题，下方「相关图标」推荐网格

#### Scenario: Related icon click navigates to its detail
- **WHEN** 用户点击推荐列表中的某个图标
- **THEN** 导航到该图标的详情页 `/workspace/icons/:new-id`

#### Scenario: No related icons
- **WHEN** 当前图标无相关推荐
- **THEN** 「相关图标」区域显示空状态提示

#### Scenario: Detail page for non-existing icon
- **WHEN** 用户访问不存在的图标 ID
- **THEN** 显示「图标不存在」提示和返回图标库按钮

#### Scenario: Private icon detail access denied
- **WHEN** 非所有者访问私有图标详情页
- **THEN** 显示「图标不存在」提示

### Requirement: Publish to library from editor
系统 SHALL 在编辑器页面提供「发布到图标库」功能，含公开/私有选择。

#### Scenario: Publish button available
- **WHEN** 编辑器中有 SVG 内容
- **THEN** 工具栏显示「发布到图标库」按钮

#### Scenario: Publish dialog
- **WHEN** 用户点击发布按钮
- **THEN** 弹窗显示：图标名称（预填）、标签输入（建议 + 自由输入）、主题选择、公开/私有勾选框，确认后调用创建 API

#### Scenario: Publish success
- **WHEN** 发布 API 返回 201
- **THEN** 显示成功 toast +「查看图标」链接跳转到详情页

### Requirement: Navbar entry
系统 SHALL 在导航栏中提供图标库入口。

#### Scenario: Navbar shows icon library link
- **WHEN** 用户已登录
- **THEN** 导航栏显示「图标库」链接，href 为 `/workspace/icons`
