import { COLORABLE_TAGS, ColorMap } from './types'

export function stripXss(svg: string): string {
  return svg
    .replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, '')
    .replace(/\son\w+\s*=\s*"[^"]*"/gi, '')
    .replace(/\son\w+\s*=\s*'[^']*'/gi, '')
}

export function parseSvg(svg: string): Document {
  const cleaned = stripXss(svg)
  const parser = new DOMParser()
  const doc = parser.parseFromString(cleaned, 'image/svg+xml')
  const errorNode = doc.querySelector('parsererror')
  if (errorNode) throw new Error('SVG 格式无效')
  return doc
}

export function buildColorMap(doc: Document): ColorMap {
  const map: ColorMap = new Map()
  for (const tag of COLORABLE_TAGS) {
    for (const el of doc.querySelectorAll(tag)) {
      for (const attr of ['fill', 'stroke'] as const) {
        const color = el.getAttribute(attr)
        if (color && color !== 'none' && color !== 'transparent') {
          if (!map.has(color)) map.set(color, new Set())
          map.get(color)!.add(el as SVGElement)
        }
      }
    }
  }
  return map
}
