# SVG 在线调色板 — 设计文档

## 概述

独立页面 `/workspace/editor`，从 Navbar 进入。粘贴或上传 SVG → 可视化渲染 → 点击选中元素 → 颜色选择器调色 → 实时替换 fill/stroke → 手动保存到 Library 或下载/复制。

纯前端实现，React 19 + TypeScript 5 + Tailwind CSS，DOMParser 解析 SVG，零重型依赖。

## 页面入口

Navbar "Editor" 链接 → `/workspace/editor`（ProtectedRoute + WorkspaceShell 包裹）

## 布局：两栏

```
┌─────────────────────┬──────────────────────┐
│                     │  ElementInspector     │
│                     │  (fill/stroke 属性)    │
│     SvgCanvas       │──────────────────────│
│     (SVG 渲染区)     │  ColorPicker          │
│                     │  · 色相条              │
│     · 粘贴区         │  · SB 面板            │
│     · 点击选元素      │  · Alpha 条           │
│     · 蓝色描边高亮    │  · 预览色块 + HEX/RGB │
│                     │──────────────────────│
│                     │  PresetColors (8色)    │
│                     │──────────────────────│
│                     │  ThemeReplacer         │
│                     │──────────────────────│
│                     │  Toolbar              │
│                     │  [撤销][重做][下载]     │
│                     │  [复制][保存到Library]  │
└─────────────────────┴──────────────────────┘
```

## 组件树

```
EditorPage
├─ SvgCanvas
│   ├─ 空状态: 拖拽/粘贴 SVG
│   ├─ 渲染状态: 可交互 SVG
│   └─ 错误状态: 格式无效
└─ SidePanel
    ├─ ElementInspector
    │   ├─ 未选中: 提示文字
    │   └─ 已选中: 元素类型 + fill/stroke 色块 + 颜色值
    ├─ FillStrokeTabs (fill / stroke 切换)
    ├─ ColorPicker
    │   ├─ HueSlider (色相条 0-360°)
    │   ├─ SBPanel (饱和度×亮度 二维面板 + 拖拽手柄)
    │   ├─ AlphaSlider (透明度 0-100%)
    │   ├─ ColorPreview (大色块 + 半透明预览)
    │   └─ ColorInput (HEX 可编辑 + RGB 只读)
    ├─ PresetColors (8 个预设，两行四列)
    ├─ ThemeReplacer (源色下拉 + 目标色 + 替换按钮)
    └─ Toolbar (撤销/重做/下载/复制/保存)
```

## 数据流

### SVG 导入 → 渲染

```
用户粘贴 / 拖拽 .svg
  → XSS strip (<script>, on* 属性)
  → DOMParser → SVGDocument
  → 提取可着色元素 (path/circle/rect/ellipse/line/polygon)
  → 绑定 click 事件 → 点击触发 onSelect(element)
  → 注入 CSS: :hover 高亮, .selected 蓝色描边
  → innerHTML 渲染到 canvas div
```

### ColorMap 构建

```typescript
type ColorMap = Map<string, Set<SVGElement>>

function buildColorMap(doc: Document): ColorMap {
  const map = new Map<string, Set<SVGElement>>()
  for (const el of doc.querySelectorAll('[fill],[stroke]')) {
    for (const attr of ['fill', 'stroke']) {
      const color = el.getAttribute(attr)
      if (color && color !== 'none' && color !== 'transparent') {
        if (!map.has(color)) map.set(color, new Set())
        map.get(color)!.add(el)
      }
    }
  }
  return map
}
```

### 换色

```
用户选中元素 + ColorPicker 产出颜色 + FillStrokeTabs 指定目标属性
  → applyColor(element, newColor, mode)
    ├─ 记录 undo: { element, oldColor, newColor, mode }
    ├─ element.setAttribute(mode, newColor)
    └─ 更新 ColorMap: 旧颜色集合移除 element，新颜色集合加入 element
```

### 全局主题替换

```
用户选源色(从ColorMap keys) + 目标色(ColorPicker)
  → themeReplace(sourceColor, targetColor, mode)
    ├─ elements = colorMap.get(sourceColor)
    ├─ elements.forEach(el => applyColor(el, targetColor, mode))
    ├─ 批量 pushUndo
    └─ colorMap.delete(sourceColor) // 旧颜色已不存在
```

### 撤销/重做

```typescript
type UndoEntry = {
  element: SVGElement
  oldColor: string | null
  newColor: string
  mode: 'fill' | 'stroke'
}

// 最多保留 50 步
const MAX_UNDO = 50

function undo() {
  const entry = undoStack.pop()
  redoStack.push(entry)
  entry.element.setAttribute(entry.mode, entry.oldColor ?? '')
  updateColorMap(entry.element, entry.newColor, entry.oldColor)
}

function redo() {
  const entry = redoStack.pop()
  undoStack.push(entry)
  entry.element.setAttribute(entry.mode, entry.newColor)
  updateColorMap(entry.element, entry.oldColor, entry.newColor)
}
```

## 颜色选择器（纯前端零依赖实现）

### 色相条 (HueSlider)
- `<input type="range" min="0" max="360">`
- CSS `background: linear-gradient(to right, #F00, #FF0, #0F0, #0FF, #00F, #F0F, #F00)`
- 值 = 当前 HSV 中的 H

### SB 面板 (Saturation × Brightness)
- 外层 `<div>` 200×200px
- 背景层: `linear-gradient(to right, #FFF, hsl(H, 100%, 50%))`
- 覆盖层: `linear-gradient(to top, #000, transparent)`
- 拖拽手柄: 圆形 `<div>`，mousedown/mousemove/mouseup
- X 轴 = Saturation (0-100%), Y 轴 = Brightness (100%-0%)

### Alpha 条
- `<input type="range" min="0" max="100">`
- 背景: checkerboard pattern + `linear-gradient(to right, transparent, currentColor)`

### 颜色转换函数（手写，约 60 行）
```typescript
hsvToRgb(h: number, s: number, v: number): [number, number, number]
rgbToHex(r: number, g: number, b: number): string
hexToRgb(hex: string): [number, number, number] | null
rgbToHsv(r: number, g: number, b: number): [number, number, number]
```

## 预设色
```typescript
const PRESETS = [
  '#EF4444', '#F97316', '#EAB308', '#22C55E',
  '#3B82F6', '#8B5CF6', '#EC4899', '#6B7280',
]
```

## 导出 / 保存

| 操作 | 行为 |
|------|------|
| 下载 SVG | `serializeToString(doc)` → Blob → `<a download>` |
| 复制代码 | `serializeToString(doc)` → `navigator.clipboard.writeText()` → toast |
| 保存到 Library | Blob → FormData → `POST /api/v1/conversions` → toast 成功/失败 |

## 安全

- 渲染前 `svgString.replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, '')`
- 渲染前 `svgString.replace(/\son\w+\s*=\s*"[^"]*"/gi, '')`
- 禁止外部资源加载（`<use href>` 等不解析）

## 状态处理

| 组件 | Loading | Empty | Error | Edge |
|------|---------|-------|-------|------|
| EditorPage | — | "粘贴 SVG 代码或拖拽 .svg 文件" | "SVG 格式无效" toast | SVG > 5000px → 缩放至视口 |
| ElementInspector | — | "点击画布中的元素查看属性" | — | 元素无 fill → 显示 "none" |
| ColorPicker | — | 默认选中第一个预设色 | HEX 格式错误 → 红色边框 + 不应用 | Alpha=0 → 预览显示 checkerboard |
| ThemeReplacer | SVG 解析中 | "暂无颜色可替换" | — | 源色=目标色 → 按钮 disabled |
| 保存到 Library | 按钮 spinner + "保存中..." | — | toast "保存失败" + 可重试 | 未登录 → 跳转登录 |
| 撤销/重做 | — | 按钮 disabled（栈空） | — | 超过 50 步 → 丢弃最旧记录 |

## 非目标 (v1)

- 不支持编辑 path 坐标点
- 不支持渐变编辑器（linearGradient/radialGradient 元素只读显示，不可编辑）
- 不修改 SVG 层级结构
- 不添加/删除 SVG 元素
- 不支持多选元素
- 不支持 eyedropper 取色

## 文件清单

```
web/src/
├─ pages/EditorPage.tsx              # 页面主组件，状态管理
├─ features/svg-editor/
│   ├─ domain/
│   │   ├─ svgParser.ts             # DOMParser + XSS strip + buildColorMap
│   │   ├─ colorUtils.ts            # HSV/RGB/HEX 互转
│   │   ├─ applyColor.ts            # applyColor + undo/redo + themeReplace
│   │   └─ types.ts                 # ColorMap, UndoEntry, PresetColor
│   ├─ components/
│   │   ├─ SvgCanvas.tsx            # SVG 渲染 + 元素点击交互
│   │   ├─ SidePanel.tsx            # 右栏容器
│   │   ├─ ElementInspector.tsx     # 元素属性显示
│   │   ├─ FillStrokeTabs.tsx       # fill/stroke 切换
│   │   ├─ ColorPicker.tsx          # 颜色选择器容器
│   │   ├─ HueSlider.tsx            # 色相条
│   │   ├─ SBPanel.tsx             # 饱和度×亮度面板
│   │   ├─ AlphaSlider.tsx          # 透明度条
│   │   ├─ ColorPreview.tsx         # 预览色块
│   │   ├─ ColorInput.tsx           # HEX/RGB 输入
│   │   ├─ PresetColors.tsx         # 8 个预设色块
│   │   ├─ ThemeReplacer.tsx        # 全局主题替换
│   │   └─ EditorToolbar.tsx        # 底部工具栏
│   └─ __tests__/
│       ├─ svgParser.test.ts
│       ├─ colorUtils.test.ts
│       ├─ applyColor.test.ts
│       ├─ ColorPicker.test.tsx
│       └─ EditorPage.test.tsx
```

修改文件:
- `web/src/App.tsx` — 添加 `/workspace/editor` 路由
- `web/src/components/Navbar.tsx` — 添加 "Editor" 链接
