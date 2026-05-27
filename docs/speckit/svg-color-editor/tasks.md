# Tasks: SVG 在线调色板

> 拆解自: docs/superpowers/specs/2026-05-27-svg-color-editor.spec.md

---

### Task 1: SVG 渲染引擎

- **描述**: SVG 字符串 → 交互式 DOM 渲染，支持元素点击高亮
- **依赖**: 无
- **估时**: 3h
- **BC**: svg-editor (前端 domain 层)

### Task 2: 元素检查器

- **描述**: 选中元素后显示 fill/stroke 属性面板
- **依赖**: Task 1
- **估时**: 2h
- **BC**: svg-editor (前端 domain 层)

### Task 3: 调色板组件

- **描述**: 8 预设色 + 自定义 hex 输入框 + 透明度滑块
- **依赖**: 无
- **估时**: 2h
- **BC**: svg-editor (前端 components)

### Task 4: 换色逻辑

- **描述**: 选区块 + 选颜色 → 替换 fill/stroke → 实时预览 + 撤销
- **依赖**: Task 1, Task 2, Task 3
- **估时**: 3h
- **BC**: svg-editor (前端 domain 层)

### Task 5: 全局主题色替换

- **描述**: 扫描 SVG 所有颜色值 → 用户选源色+目标色 → 一键全局替换
- **依赖**: Task 1, Task 3
- **估时**: 2h
- **BC**: svg-editor (前端 domain 层)

### Task 6: 导入/导出

- **描述**: 粘贴 SVG / 上传文件导入，下载 / 复制 / 保存到 Library
- **依赖**: Task 1
- **估时**: 2h
- **BC**: svg-editor (前端 components)

### Task 7: 页面路由 + 布局

- **描述**: /editor 路由注册、三栏布局框架（左侧面板 + 中间画布 + 右侧调色板）
- **依赖**: Task 1, Task 2, Task 3
- **估时**: 2h
- **BC**: svg-editor (前端 pages + components)
