import { describe, it, expect, beforeEach } from 'vitest'
import { createColorState, applyColor, undo, redo, themeReplace } from '../domain/applyColor'
import { parseSvg } from '../domain/svgParser'

function makeDoc() {
  return parseSvg(`<svg xmlns="http://www.w3.org/2000/svg">
    <rect id="r1" fill="#FF0000" />
    <rect id="r2" fill="#FF0000" />
    <circle id="c1" fill="#0000FF" />
  </svg>`)
}

describe('createColorState', () => {
  it('从 SVG 文档构建状态', () => {
    const doc = makeDoc()
    const state = createColorState(doc)
    expect(state.colorMap.get('#FF0000')?.size).toBe(2)
  })
})

describe('applyColor', () => {
  let state: ReturnType<typeof createColorState>

  beforeEach(() => {
    state = createColorState(makeDoc())
  })

  it('修改元素填充颜色', () => {
    const el = state.doc.getElementById('r1')! as unknown as SVGElement
    applyColor(state, el, '#00FF00', 'fill')
    expect(el.getAttribute('fill')).toBe('#00FF00')
    expect(state.undoStack.length).toBe(1)
  })

  it('应用颜色后更新 colorMap', () => {
    const el = state.doc.getElementById('r1')! as unknown as SVGElement
    applyColor(state, el, '#00FF00', 'fill')
    expect(state.colorMap.get('#FF0000')?.size).toBe(1) // r2 仍然是红色
    expect(state.colorMap.get('#00FF00')?.size).toBe(1) // r1 现在是绿色
  })
})

describe('undo / redo', () => {
  let state: ReturnType<typeof createColorState>

  beforeEach(() => {
    state = createColorState(makeDoc())
    const el = state.doc.getElementById('r1')! as unknown as SVGElement
    applyColor(state, el, '#00FF00', 'fill')
  })

  it('撤销恢复之前的颜色', () => {
    const el = state.doc.getElementById('r1')! as unknown as SVGElement
    undo(state)
    expect(el.getAttribute('fill')).toBe('#FF0000')
    expect(state.redoStack.length).toBe(1)
  })

  it('重做重新应用更改', () => {
    const el = state.doc.getElementById('r1')!
    undo(state)
    redo(state)
    expect(el.getAttribute('fill')).toBe('#00FF00')
  })

  it('撤销后再应用会清空重做栈', () => {
    const el = state.doc.getElementById('r1')! as unknown as SVGElement
    undo(state)
    applyColor(state, el, '#000000', 'fill')
    expect(state.redoStack.length).toBe(0)
  })
})

describe('themeReplace', () => {
  it('替换所有匹配的颜色', () => {
    const state = createColorState(makeDoc())
    themeReplace(state, '#FF0000', '#000000', 'fill')
    expect(state.doc.getElementById('r1')!.getAttribute('fill')).toBe('#000000')
    expect(state.doc.getElementById('r2')!.getAttribute('fill')).toBe('#000000')
    expect(state.colorMap.has('#FF0000')).toBe(false)
  })

  it('源颜色不存在时不做任何操作', () => {
    const state = createColorState(makeDoc())
    themeReplace(state, '#BADBAD', '#000000', 'fill')
    expect(state.undoStack.length).toBe(0)
  })
})

describe('MAX_UNDO', () => {
  it('栈超过 50 时丢弃最早的记录', () => {
    const state = createColorState(makeDoc())
    const el = state.doc.getElementById('r1')! as unknown as SVGElement
    for (let i = 0; i < 55; i++) {
      applyColor(state, el, `#00000${(i % 10)}`, 'fill')
    }
    expect(state.undoStack.length).toBe(50)
  })
})
