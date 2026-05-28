/** Convert HSV (h: 0-360, s: 0-100, v: 0-100) to RGB (0-255 each) */
export function hsvToRgb(h: number, s: number, v: number): [number, number, number] {
  const sNorm = s / 100
  const vNorm = v / 100
  const c = vNorm * sNorm
  const x = c * (1 - Math.abs(((h / 60) % 2) - 1))
  const m = vNorm - c

  let r = 0, g = 0, b = 0
  if (h < 60) { r = c; g = x; b = 0 }
  else if (h < 120) { r = x; g = c; b = 0 }
  else if (h < 180) { r = 0; g = c; b = x }
  else if (h < 240) { r = 0; g = x; b = c }
  else if (h < 300) { r = x; g = 0; b = c }
  else { r = c; g = 0; b = x }

  return [
    Math.round((r + m) * 255),
    Math.round((g + m) * 255),
    Math.round((b + m) * 255),
  ]
}

/** Convert RGB (0-255 each) to HEX string (#RRGGBB) */
export function rgbToHex(r: number, g: number, b: number): string {
  const toHex = (n: number) => Math.max(0, Math.min(255, n)).toString(16).padStart(2, '0').toUpperCase()
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`
}

/** Parse HEX (#RRGGBB or #RGB) to RGB, returns null for invalid input */
export function hexToRgb(hex: string): [number, number, number] | null {
  const match = hex.match(/^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$/)
  if (!match) return null
  let h = match[1]
  if (h.length === 3) h = h[0] + h[0] + h[1] + h[1] + h[2] + h[2]
  return [
    parseInt(h.slice(0, 2), 16),
    parseInt(h.slice(2, 4), 16),
    parseInt(h.slice(4, 6), 16),
  ]
}

/** Convert RGB (0-255 each) to HSV (h: 0-360, s: 0-100, v: 0-100) */
export function rgbToHsv(r: number, g: number, b: number): [number, number, number] {
  const rNorm = r / 255
  const gNorm = g / 255
  const bNorm = b / 255
  const max = Math.max(rNorm, gNorm, bNorm)
  const min = Math.min(rNorm, gNorm, bNorm)
  const delta = max - min

  let h = 0
  if (delta !== 0) {
    if (max === rNorm) h = 60 * (((gNorm - bNorm) / delta) % 6)
    else if (max === gNorm) h = 60 * (((bNorm - rNorm) / delta) + 2)
    else h = 60 * (((rNorm - gNorm) / delta) + 4)
  }
  if (h < 0) h += 360

  const s = max === 0 ? 0 : (delta / max) * 100
  const v = max * 100

  return [Math.round(h), Math.round(s), Math.round(v)]
}
