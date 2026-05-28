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
  { hex: '#EF4444', name: 'Red' },
  { hex: '#F97316', name: 'Orange' },
  { hex: '#EAB308', name: 'Yellow' },
  { hex: '#22C55E', name: 'Green' },
  { hex: '#3B82F6', name: 'Blue' },
  { hex: '#8B5CF6', name: 'Purple' },
  { hex: '#EC4899', name: 'Pink' },
  { hex: '#6B7280', name: 'Gray' },
]

export const COLORABLE_TAGS = new Set([
  'path', 'circle', 'rect', 'ellipse', 'line', 'polygon', 'polyline', 'text', 'g',
])
