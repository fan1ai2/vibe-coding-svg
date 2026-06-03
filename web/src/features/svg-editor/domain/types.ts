export type ColorMode = 'fill' | 'stroke'

export type ColorMap = Map<string, Set<SVGElement>>

export interface UndoEntry {
  element: SVGElement
  oldColor: string | null
  newColor: string
  mode: ColorMode
}

export interface PresetColor {
  hex: string
  name: string
}

export const PRESETS: PresetColor[] = [
  { hex: '#EF4444', name: '红色' },
  { hex: '#F97316', name: '橙色' },
  { hex: '#EAB308', name: '黄色' },
  { hex: '#22C55E', name: '绿色' },
  { hex: '#3B82F6', name: '蓝色' },
  { hex: '#8B5CF6', name: '紫色' },
  { hex: '#EC4899', name: '粉色' },
  { hex: '#6B7280', name: '灰色' },
]

export const COLORABLE_TAGS = new Set([
  'path', 'circle', 'rect', 'ellipse', 'line', 'polygon', 'polyline', 'text', 'g',
])
