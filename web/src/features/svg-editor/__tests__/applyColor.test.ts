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
  it('builds state from SVG doc', () => {
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

  it('changes element fill color', () => {
    const el = state.doc.getElementById('r1')!
    applyColor(state, el, '#00FF00', 'fill')
    expect(el.getAttribute('fill')).toBe('#00FF00')
    expect(state.undoStack.length).toBe(1)
  })

  it('updates colorMap after apply', () => {
    const el = state.doc.getElementById('r1')!
    applyColor(state, el, '#00FF00', 'fill')
    expect(state.colorMap.get('#FF0000')?.size).toBe(1) // r2 still red
    expect(state.colorMap.get('#00FF00')?.size).toBe(1) // r1 now green
  })
})

describe('undo / redo', () => {
  let state: ReturnType<typeof createColorState>

  beforeEach(() => {
    state = createColorState(makeDoc())
    const el = state.doc.getElementById('r1')!
    applyColor(state, el, '#00FF00', 'fill')
  })

  it('undo restores previous color', () => {
    const el = state.doc.getElementById('r1')!
    undo(state)
    expect(el.getAttribute('fill')).toBe('#FF0000')
    expect(state.redoStack.length).toBe(1)
  })

  it('redo re-applies the change', () => {
    const el = state.doc.getElementById('r1')!
    undo(state)
    redo(state)
    expect(el.getAttribute('fill')).toBe('#00FF00')
  })

  it('undo then apply clears redo stack', () => {
    const el = state.doc.getElementById('r1')!
    undo(state)
    applyColor(state, el, '#000000', 'fill')
    expect(state.redoStack.length).toBe(0)
  })
})

describe('themeReplace', () => {
  it('replaces all occurrences of a color', () => {
    const state = createColorState(makeDoc())
    themeReplace(state, '#FF0000', '#000000', 'fill')
    expect(state.doc.getElementById('r1')!.getAttribute('fill')).toBe('#000000')
    expect(state.doc.getElementById('r2')!.getAttribute('fill')).toBe('#000000')
    expect(state.colorMap.has('#FF0000')).toBe(false)
  })

  it('does nothing when source color does not exist', () => {
    const state = createColorState(makeDoc())
    themeReplace(state, '#BADBAD', '#000000', 'fill')
    expect(state.undoStack.length).toBe(0)
  })
})

describe('MAX_UNDO', () => {
  it('discards oldest entry when stack exceeds 50', () => {
    const state = createColorState(makeDoc())
    const el = state.doc.getElementById('r1')!
    for (let i = 0; i < 55; i++) {
      applyColor(state, el, `#00000${(i % 10)}`, 'fill')
    }
    expect(state.undoStack.length).toBe(50)
  })
})
