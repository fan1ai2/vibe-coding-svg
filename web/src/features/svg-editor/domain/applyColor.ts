import { ColorMap, ColorMode, UndoEntry } from './types'
import { buildColorMap } from './svgParser'

const MAX_UNDO = 50

export interface ColorState {
  doc: Document
  colorMap: ColorMap
  undoStack: UndoEntry[]
  redoStack: UndoEntry[]
}

export function createColorState(doc: Document): ColorState {
  return {
    doc,
    colorMap: buildColorMap(doc),
    undoStack: [],
    redoStack: [],
  }
}

function updateColorMapEntry(map: ColorMap, color: string | null, element: SVGElement, add: boolean) {
  if (!color || color === 'none' || color === 'transparent') return
  if (add) {
    if (!map.has(color)) map.set(color, new Set())
    map.get(color)!.add(element)
  } else {
    const set = map.get(color)
    if (set) {
      set.delete(element)
      if (set.size === 0) map.delete(color)
    }
  }
}

export function applyColor(
  state: ColorState,
  element: SVGElement,
  newColor: string,
  mode: ColorMode,
) {
  const oldColor = element.getAttribute(mode)

  state.undoStack.push({ element, oldColor, newColor, mode })
  if (state.undoStack.length > MAX_UNDO) state.undoStack.shift()
  state.redoStack = []

  element.setAttribute(mode, newColor)
  updateColorMapEntry(state.colorMap, oldColor, element, false)
  updateColorMapEntry(state.colorMap, newColor, element, true)
}

export function undo(state: ColorState) {
  const entry = state.undoStack.pop()
  if (!entry) return
  state.redoStack.push(entry)

  entry.element.setAttribute(entry.mode, entry.oldColor ?? '')
  updateColorMapEntry(state.colorMap, entry.newColor, entry.element, false)
  updateColorMapEntry(state.colorMap, entry.oldColor, entry.element, true)
}

export function redo(state: ColorState) {
  const entry = state.redoStack.pop()
  if (!entry) return
  state.undoStack.push(entry)

  entry.element.setAttribute(entry.mode, entry.newColor)
  updateColorMapEntry(state.colorMap, entry.oldColor, entry.element, false)
  updateColorMapEntry(state.colorMap, entry.newColor, entry.element, true)
}

export function themeReplace(
  state: ColorState,
  sourceColor: string,
  targetColor: string,
  mode: ColorMode,
) {
  const elements = state.colorMap.get(sourceColor)
  if (!elements) return
  const snapshot = Array.from(elements)
  for (const el of snapshot) {
    applyColor(state, el, targetColor, mode)
  }
}
