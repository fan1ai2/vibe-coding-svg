import { describe, it, expect } from 'vitest'
import { parseSvg, buildColorMap, stripXss } from '../domain/svgParser'

const sampleSvg = `<svg xmlns="http://www.w3.org/2000/svg">
  <rect fill="#FF0000" />
  <circle fill="#FF0000" stroke="#0000FF" />
  <path fill="none" />
  <rect fill="#00FF00" />
</svg>`

describe('stripXss', () => {
  it('removes script tags', () => {
    const input = '<svg><script>alert("xss")</script><rect/></svg>'
    const result = stripXss(input)
    expect(result).not.toContain('script')
    expect(result).toContain('<rect')
  })
  it('removes on* event attributes', () => {
    const input = '<circle onclick="alert(1)" fill="#FF0000"/>'
    const result = stripXss(input)
    expect(result).not.toContain('onclick')
    expect(result).toContain('#FF0000')
  })
})

describe('parseSvg', () => {
  it('parses valid SVG string to Document', () => {
    const doc = parseSvg(sampleSvg)
    expect(doc.querySelectorAll('rect').length).toBe(2)
    expect(doc.querySelectorAll('circle').length).toBe(1)
  })
  it('throws on invalid SVG', () => {
    expect(() => parseSvg('not an svg')).toThrow()
  })
})

describe('buildColorMap', () => {
  it('maps colors to elements', () => {
    const doc = parseSvg(sampleSvg)
    const map = buildColorMap(doc)
    expect(map.get('#FF0000')?.size).toBe(2)
    expect(map.get('#0000FF')?.size).toBe(1)
    expect(map.get('#00FF00')?.size).toBe(1)
  })
  it('excludes "none" and "transparent"', () => {
    const doc = parseSvg(sampleSvg)
    const map = buildColorMap(doc)
    expect(map.has('none')).toBe(false)
  })
})
