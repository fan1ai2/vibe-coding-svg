## ADDED Requirements

### Requirement: Recommend related icons with weighted scoring
系统 SHALL 基于 Neo4j 图关系返回与指定图标共享标签、颜色或主题的相关图标，按标签类型加权评分排序。

#### Scenario: Related by shared tags with type weighting
- **WHEN** 图标 A 有 usage 标签 [home]、style 标签 [flat]；图标 B 有 usage 标签 [home]、style 标签 [line]；图标 C 仅有 style 标签 [flat]
- **THEN** B 得分 3（共享 usage home），C 得分 2（共享 style flat），B 排在 C 前面

#### Scenario: Related by shared tags and colors
- **WHEN** 图标 A 有 usage 标签 [home] 和颜色 [#FF0000]；图标 B 有 usage 标签 [home]、style 标签 [flat] 和颜色 [#FF0000]；图标 C 仅有 usage 标签 [home]
- **THEN** B (3+2+1=6) 排在 C (3) 前面

#### Scenario: No related icons
- **WHEN** 请求的图标在 Neo4j 中无任何共享关系的其他公开图标
- **THEN** 返回空数组 `[]`

#### Scenario: Recommend limit
- **WHEN** 请求 `GET /icons/:id/recommend?limit=5`
- **THEN** 最多返回 5 条相关图标

#### Scenario: Only public icons in recommendations
- **WHEN** 某个关联图标 is_public=false
- **THEN** 该图标不出现在推荐结果中

### Requirement: Recommend results include metadata
系统 SHALL 在推荐结果中附带每个推荐图标的加权得分和共享标签列表。

#### Scenario: Result with score and tags
- **WHEN** 推荐接口返回图标 B
- **THEN** 响应中包含 `score: 6` 和 `shared_tags: ["home", "flat"]`，附完整 SVG 内容供前端渲染缩略图

### Requirement: Neo4j sync on icon creation
系统 SHALL 在图标创建后异步在 Neo4j 中建立对应的节点和关系，Tag 节点包含 type 属性。

#### Scenario: New icon sync creates node and relationships with tag type
- **WHEN** 图标创建成功且带有 tags=[{"name":"home","type":"usage"}]、colors=["#FF0000"]、theme="business"
- **THEN** Neo4j Tag 节点存储 `{name:"home", slug:"home", type:"usage"}`，HAS_TAG/HAS_COLOR/IN_THEME 关系被建立

### Requirement: Neo4j sync on icon deletion
系统 SHALL 在图标删除后异步移除 Neo4j 中的对应节点和关系。

#### Scenario: Icon deletion cleans up graph
- **WHEN** 图标被删除
- **THEN** Neo4j 中该 Icon 节点的所有关系被删除，Icon 节点本身被删除，孤立节点保留以便复用
