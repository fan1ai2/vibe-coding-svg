import { describe, it, expect } from 'vitest'
import { parseSvg, buildColorMap, stripXss } from '../domain/svgParser'

const sampleSvg = `<svg xmlns="http://www.w3.org/2000/svg">
  <rect fill="#FF0000" />
  <circle fill="#FF0000" stroke="#0000FF" />
  <path fill="none" />
  <rect fill="#00FF00" />
</svg>`

describe('stripXss', () => {
  it('移除 script 标签', () => {
    const input = '<svg><script>alert("xss")</script><rect/></svg>'
    const result = stripXss(input)
    expect(result).not.toContain('script')
    expect(result).toContain('<rect')
  })
  it('移除 on* 事件属性', () => {
    const input = '<circle onclick="alert(1)" fill="#FF0000"/>'
    const result = stripXss(input)
    expect(result).not.toContain('onclick')
    expect(result).toContain('#FF0000')
  })
})

describe('parseSvg', () => {
  it('将有效 SVG 字符串解析为 Document', () => {
    const doc = parseSvg(sampleSvg)
    expect(doc.querySelectorAll('rect').length).toBe(2)
    expect(doc.querySelectorAll('circle').length).toBe(1)
  })
  it('无效 SVG 抛出异常', () => {
    expect(() => parseSvg('not an svg')).toThrow()
  })
})

describe('buildColorMap', () => {
  it('将颜色映射到对应元素', () => {
    const doc = parseSvg(sampleSvg)
    const map = buildColorMap(doc)
    expect(map.get('#FF0000')?.size).toBe(2)
    expect(map.get('#0000FF')?.size).toBe(1)
    expect(map.get('#00FF00')?.size).toBe(1)
  })
  it('排除 "none" 和 "transparent"', () => {
    const doc = parseSvg(sampleSvg)
    const map = buildColorMap(doc)
    expect(map.has('none')).toBe(false)
  })
})
