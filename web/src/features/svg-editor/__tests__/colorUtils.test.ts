import { describe, it, expect } from 'vitest'
import { hsvToRgb, rgbToHex, hexToRgb, rgbToHsv } from '../domain/colorUtils'

describe('hsvToRgb', () => {
  it('converts red (0, 100, 100) to [255, 0, 0]', () => {
    expect(hsvToRgb(0, 100, 100)).toEqual([255, 0, 0])
  })
  it('converts black (0, 0, 0) to [0, 0, 0]', () => {
    expect(hsvToRgb(0, 0, 0)).toEqual([0, 0, 0])
  })
  it('converts white (0, 0, 100) to [255, 255, 255]', () => {
    expect(hsvToRgb(0, 0, 100)).toEqual([255, 255, 255])
  })
  it('converts green (120, 100, 100) to [0, 255, 0]', () => {
    expect(hsvToRgb(120, 100, 100)).toEqual([0, 255, 0])
  })
})

describe('rgbToHex', () => {
  it('converts [255, 0, 0] to #FF0000', () => {
    expect(rgbToHex(255, 0, 0)).toBe('#FF0000')
  })
  it('converts [0, 0, 0] to #000000', () => {
    expect(rgbToHex(0, 0, 0)).toBe('#000000')
  })
  it('pads single-digit hex values', () => {
    expect(rgbToHex(1, 2, 3)).toBe('#010203')
  })
})

describe('hexToRgb', () => {
  it('parses #FF0000 to [255, 0, 0]', () => {
    expect(hexToRgb('#FF0000')).toEqual([255, 0, 0])
  })
  it('parses #ff0000 (lowercase)', () => {
    expect(hexToRgb('#ff0000')).toEqual([255, 0, 0])
  })
  it('returns null for invalid hex', () => {
    expect(hexToRgb('not-a-color')).toBeNull()
    expect(hexToRgb('#GGG')).toBeNull()
  })
})

describe('rgbToHsv', () => {
  it('converts [255, 0, 0] to [0, 100, 100]', () => {
    const [h, s, v] = rgbToHsv(255, 0, 0)
    expect(h).toBe(0)
    expect(s).toBe(100)
    expect(v).toBe(100)
  })
  it('converts [0, 0, 0] to [0, 0, 0]', () => {
    const [, s, v] = rgbToHsv(0, 0, 0)
    expect(s).toBe(0)
    expect(v).toBe(0)
  })
})

describe('roundtrip', () => {
  it('hsv → rgb → hsv preserves values', () => {
    const [h, s, v] = rgbToHsv(...hsvToRgb(200, 75, 50))
    expect(h).toBeCloseTo(200, 0)
    expect(s).toBeCloseTo(75, 0)
    expect(v).toBeCloseTo(50, 0)
  })
})
