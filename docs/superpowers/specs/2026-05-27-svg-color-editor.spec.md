# SVG 在线调色板

## 概述

独立页面 `/editor`，支持粘贴/上传 SVG → 可视化选区块 → 一键换色 → 全局主题色替换 → 导出。

## 功能列表

### 1. SVG 导入
- 粘贴 SVG 代码文本（支持 Ctrl+V 直接粘贴到画布区域）
- 上传 .svg 文件（支持拖拽到画布区域）
- 渲染前 strip `<script>` 和 `on*` 事件属性（XSS 防护）
- SVG 尺寸 > 5000px 时等比缩放至视口内
- SVG 格式非法时 toast 提示 "SVG 格式无效"，不渲染

### 2. 区块选择
- 点击 SVG 元素（`<path>` / `<circle>` / `<rect>` 等）高亮选中
- 选中后显示当前 fill 和 stroke 属性值
- 点击空白区域取消选中

### 3. 调色板
- 8 个预设主题色块
- 自定义 hex 颜色输入框（支持 #RRGGBB 格式）
- 透明度滑块（alpha 0-100%）
- 点击色块 = 应用颜色到选中元素

### 4. 一键换色
- 选中元素 + 选颜色 → 实时替换 fill（默认）
- 用户可切换为替换 stroke
- 支持 Ctrl+Z 撤销最近 20 次换色操作
- 无 fill 属性的元素默认补 `fill="currentColor"`

### 5. 全局主题色替换
- 扫描 SVG 中所有出现过的颜色值
- 用户选择源色 → 选择目标色 → 一键全局替换
- 仅替换完全匹配的颜色值

### 6. 导出
- 下载为 .svg 文件（文件名 = `edited-{timestamp}.svg`）
- 复制 SVG 代码到剪贴板
- 保存到 Library（需登录态）

## 技术约束

- 纯前端实现，不涉及额外后端 API（保存到 Library 复用现有接口）
- 不引入重型编辑器框架 — 用 DOMParser + 原生 DOM API
- 使用现有 React 19 + TypeScript 5 + Tailwind CSS 栈
- 路由: `/editor`，React Router 管理

## 非目标（v1 不做）

- 不支持编辑 path 坐标点（不做矢量路径编辑）
- 不支持渐变编辑器（v1 只做纯色）
- 不修改 SVG 层级/DOM 结构
- 不添加/删除 SVG 元素

## 安全

- SVG 渲染前必须 strip `<script>` 标签和 `on*` 事件处理器
- 不执行内联 JavaScript
- 外部资源引用（`<use href="...">`）不解析
