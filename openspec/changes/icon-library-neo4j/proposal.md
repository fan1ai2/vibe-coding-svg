## Why

当前 SVG 资源工坊只能做位图转矢量、编辑器改颜色，但没有图标资产的存储、组织和检索能力。用户需要一套图标库系统：把编辑好的 SVG 图标存进去、打标签、按颜色归类，并且能在检索时看到相关联的其他图标。这是一个从「单次转换工具」到「图标资产管理平台」的关键升级。

## What Changes

- 新增图标库数据模型：PostgreSQL 存储图标元数据 + SVG 内容，Neo4j 管理标签/颜色/主题/图标间的网状关系
- 新增图标 CRUD API：创建、列表、详情、删除，支持分页
- 新增图标检索 API：关键词搜索 + 标签筛选 + 颜色过滤 + 排序
- 新增关联推荐 API：基于 Neo4j 图查询，返回与当前图标共享属性的相关图标
- 新增图标库前端页面：搜索栏 + 标签云 + 颜色筛选器 + 图标网格 + 关联推荐列表
- 新增图标详情页：大图预览 + 元信息编辑 + 关联推荐展示
- 编辑器集成「发布到图标库」按钮，自动提取 SVG 颜色
- docker-compose 新增 Neo4j 容器

## Capabilities

### New Capabilities
- `icon-storage`: 图标数据持久化，PostgreSQL 存储元数据和 SVG 内容，支持批量入库
- `icon-search`: 多维度图标检索 —— 关键词、标签、颜色、主题组合筛选，支持排序
- `icon-graph-recommend`: Neo4j 图关系网络，基于共享标签/颜色/主题的关联图标推荐
- `icon-library-ui`: 图标库浏览页面（搜索 + 筛选 + 网格 + 分页）和图标详情页面

### Modified Capabilities
<!-- 无已有规格需要修改 -->

## Impact

- **Infrastructure**: docker-compose 新增 Neo4j 容器 (neo4j:5-community, 暴露 7474/7687 端口)，新增 neo4jdata volume
- **Database**: PostgreSQL 新增 icons / tags / icon_tags / icon_colors / icon_themes 表
- **Dependencies**: Go 服务新增 `github.com/neo4j/neo4j-go-driver/v5`
- **API**: 新增 `/api/v1/icons` 路由组 (CRUD + search + recommend)
- **Frontend**: 新增 IconLibraryPage、IconDetailPage，Navbar 加图标库入口，EditorPage 加发布按钮
