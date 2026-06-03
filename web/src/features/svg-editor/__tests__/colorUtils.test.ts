import { describe, it, expect } from 'vitest'
import { hsvToRgb, rgbToHex, hexToRgb, rgbToHsv } from '../domain/colorUtils'

describe('hsvToRgb', () => {
  it('将红色 (0, 100, 100) 转换为 [255, 0, 0]', () => {
    expect(hsvToRgb(0, 100, 100)).toEqual([255, 0, 0])
  })
  it('将黑色 (0, 0, 0) 转换为 [0, 0, 0]', () => {
    expect(hsvToRgb(0, 0, 0)).toEqual([0, 0, 0])
  })
  it('将白色 (0, 0, 100) 转换为 [255, 255, 255]', () => {
    expect(hsvToRgb(0, 0, 100)).toEqual([255, 255, 255])
  })
  it('将绿色 (120, 100, 100) 转换为 [0, 255, 0]', () => {
    expect(hsvToRgb(120, 100, 100)).toEqual([0, 255, 0])
  })
})

describe('rgbToHex', () => {
  it('将 [255, 0, 0] 转换为 #FF0000', () => {
    expect(rgbToHex(255, 0, 0)).toBe('#FF0000')
  })
  it('将 [0, 0, 0] 转换为 #000000', () => {
    expect(rgbToHex(0, 0, 0)).toBe('#000000')
  })
  it('单个数字的 hex 值正确补零', () => {
    expect(rgbToHex(1, 2, 3)).toBe('#010203')
  })
})

describe('hexToRgb', () => {
  it('将 #FF0000 解析为 [255, 0, 0]', () => {
    expect(hexToRgb('#FF0000')).toEqual([255, 0, 0])
  })
  it('将 #ff0000（小写）正确解析', () => {
    expect(hexToRgb('#ff0000')).toEqual([255, 0, 0])
  })
  it('无效 hex 返回 null', () => {
    expect(hexToRgb('not-a-color')).toBeNull()
    expect(hexToRgb('#GGG')).toBeNull()
  })
})

describe('rgbToHsv', () => {
  it('将 [255, 0, 0] 转换为 [0, 100, 100]', () => {
    const [h, s, v] = rgbToHsv(255, 0, 0)
    expect(h).toBe(0)
    expect(s).toBe(100)
    expect(v).toBe(100)
  })
  it('将 [0, 0, 0] 转换为 [0, 0, 0]', () => {
    const [, s, v] = rgbToHsv(0, 0, 0)
    expect(s).toBe(0)
    expect(v).toBe(0)
  })
})

describe('往返转换', () => {
  it('hsv → rgb → hsv 保持数值不变', () => {
    const [h, s, v] = rgbToHsv(...hsvToRgb(200, 75, 50))
    expect(h).toBeCloseTo(200, 0)
    expect(s).toBeCloseTo(75, 0)
    expect(v).toBeCloseTo(50, 0)
  })
})
