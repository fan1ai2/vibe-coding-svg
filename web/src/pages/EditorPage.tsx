import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { conversions } from '../api/client'
import { createColorState, applyColor, undo, redo, themeReplace } from '../features/svg-editor/domain/applyColor'
import { ColorMode } from '../features/svg-editor/domain/types'
import SvgCanvas from '../features/svg-editor/components/SvgCanvas'
import SidePanel from '../features/svg-editor/components/SidePanel'
import ElementInspector from '../features/svg-editor/components/ElementInspector'
import FillStrokeTabs from '../features/svg-editor/components/FillStrokeTabs'
import ColorPicker from '../features/svg-editor/components/ColorPicker'
import PresetColors from '../features/svg-editor/components/PresetColors'
import ThemeReplacer from '../features/svg-editor/components/ThemeReplacer'
import EditorToolbar from '../features/svg-editor/components/EditorToolbar'

export default function EditorPage() {
  const { token } = useAuth()
  const navigate = useNavigate()

  const [svgString, setSvgString] = useState<string | null>(null)
  const [svgDoc, setSvgDoc] = useState<Document | null>(null)
  const [colorState, setColorState] = useState<ReturnType<typeof createColorState> | null>(null)
  const [selectedElement, setSelectedElement] = useState<SVGElement | null>(null)
  const [currentColor, setCurrentColor] = useState('#3B82F6')
  const [alpha, setAlpha] = useState(100)
  const [mode, setMode] = useState<ColorMode>('fill')
  const [toast, setToast] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [renderTick, setRenderTick] = useState(0)

  const showToast = useCallback((msg: string) => {
    setToast(msg)
    setTimeout(() => setToast(null), 3000)
  }, [])

  const handleSvgLoaded = useCallback((svg: string, doc: Document) => {
    setSvgString(svg)
    setSvgDoc(doc)
    const state = createColorState(doc)
    setColorState(state)
  }, [])

  const handleElementSelect = useCallback((el: SVGElement | null) => {
    setSelectedElement(el)
  }, [])

  const handleApplyColor = useCallback(() => {
    if (!selectedElement || !colorState) return
    applyColor(colorState, selectedElement, currentColor, mode)
    setRenderTick(n => n + 1)
  }, [selectedElement, currentColor, mode, colorState])

  const handleUndo = useCallback(() => {
    if (!colorState || colorState.undoStack.length === 0) return
    undo(colorState)
    setRenderTick(n => n + 1)
  }, [colorState])

  const handleRedo = useCallback(() => {
    if (!colorState || colorState.redoStack.length === 0) return
    redo(colorState)
    setRenderTick(n => n + 1)
  }, [colorState])

  const handleThemeReplace = useCallback((sourceColor: string, targetColor: string, m: ColorMode) => {
    if (!colorState) return
    themeReplace(colorState, sourceColor, targetColor, m)
    setRenderTick(n => n + 1)
    showToast(`已将所有 ${sourceColor} 替换为 ${targetColor}`)
  }, [colorState, showToast])

  const serializeSvg = useCallback((): string | null => {
    if (!svgDoc) return null
    const serializer = new XMLSerializer()
    return serializer.serializeToString(svgDoc.documentElement)
  }, [svgDoc])

  const handleDownload = useCallback(() => {
    const str = serializeSvg()
    if (!str) return
    const blob = new Blob([str], { type: 'image/svg+xml' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `edited-${Date.now()}.svg`
    a.click()
    URL.revokeObjectURL(url)
    showToast('下载完成')
  }, [serializeSvg, showToast])

  const handleCopy = useCallback(async () => {
    const str = serializeSvg()
    if (!str) return
    await navigator.clipboard.writeText(str)
    showToast('已复制到剪贴板')
  }, [serializeSvg, showToast])

  const handleSave = useCallback(async () => {
    if (!token) { navigate('/'); return }
    const str = serializeSvg()
    if (!str) return
    const blob = new Blob([str], { type: 'image/svg+xml' })
    const file = new File([blob], `edited-${Date.now()}.svg`, { type: 'image/svg+xml' })
    try {
      await conversions.upload(file)
      showToast('已保存到 Library')
    } catch {
      setError('保存失败，请重试')
      setTimeout(() => setError(null), 3000)
    }
  }, [token, navigate, showToast, serializeSvg])

  return (
    <div className="flex flex-col gap-4 h-full">
      {toast && (
        <div className="fixed top-20 right-4 z-50 px-4 py-2 bg-gray-800 text-white text-sm rounded-lg shadow-lg">
          {toast}
        </div>
      )}
      {error && (
        <div className="fixed top-20 right-4 z-50 px-4 py-2 bg-red-500 text-white text-sm rounded-lg shadow-lg">
          {error}
        </div>
      )}

      <div className="flex gap-4 flex-1 min-h-0">
        <SvgCanvas
          svg={svgString}
          selectedElement={selectedElement}
          onSelect={handleElementSelect}
          onError={setError}
          onSvgLoaded={handleSvgLoaded}
        />
        <SidePanel>
          <ElementInspector element={selectedElement} />
          <FillStrokeTabs mode={mode} onChange={setMode} />
          <ColorPicker
            color={currentColor}
            alpha={alpha}
            onColorChange={setCurrentColor}
            onAlphaChange={setAlpha}
          />
          <PresetColors onSelect={(hex) => {
            setCurrentColor(hex)
            if (selectedElement && colorState) {
              applyColor(colorState, selectedElement, hex, mode)
              setRenderTick(n => n + 1)
            }
          }} />
          {colorState && (
            <ThemeReplacer
              colorMap={colorState.colorMap}
              targetColor={currentColor}
              mode={mode}
              onReplace={handleThemeReplace}
            />
          )}
          <EditorToolbar
            canUndo={!!colorState && colorState.undoStack.length > 0}
            canRedo={!!colorState && colorState.redoStack.length > 0}
            canExport={svgString !== null}
            onUndo={handleUndo}
            onRedo={handleRedo}
            onDownload={handleDownload}
            onCopy={handleCopy}
            onSave={handleSave}
          />
        </SidePanel>
      </div>
    </div>
  )
}
