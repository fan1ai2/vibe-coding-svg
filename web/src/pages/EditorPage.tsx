import { useState, useCallback, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { svgs, icons, ApiError } from '../api/client'
import PublishDialog from '../components/PublishDialog'
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
  const [colorState, setColorState] = useState<ReturnType<typeof createColorState> | null>(null)
  const [selectedElement, setSelectedElement] = useState<SVGElement | null>(null)
  const [currentColor, setCurrentColor] = useState('#3B82F6')
  const [alpha, setAlpha] = useState(100)
  const [mode, setMode] = useState<ColorMode>('fill')
  const [toast, setToast] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [showPublish, setShowPublish] = useState(false)
  const [renderTick, setRenderTick] = useState(0)
  void renderTick

  const showToast = useCallback((msg: string) => {
    setToast(msg)
    setTimeout(() => setToast(null), 3000)
  }, [])

  // Load SVG passed via sessionStorage (from conversion preview or icon detail)
  useEffect(() => {
    const pending = sessionStorage.getItem('editor:pendingSvg')
    if (pending) {
      sessionStorage.removeItem('editor:pendingSvg')
      setSvgString(pending)
    }
  }, [])

  const handleSvgLoaded = useCallback((svg: string, doc: Document) => {
    setSvgString(svg)
    const state = createColorState(doc)
    setColorState(state)
  }, [])

  const handleElementSelect = useCallback((el: SVGElement | null) => {
    setSelectedElement(el)
  }, [])

  const applyAndTick = useCallback((el: SVGElement | null, color: string, m: ColorMode) => {
    if (!el || !colorState) return
    applyColor(colorState, el, color, m)
    setRenderTick(n => n + 1)
  }, [colorState])

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
    const svgEl = document.querySelector('[data-editor-svg]') as SVGElement | null
    if (!svgEl) return null
    const clone = svgEl.cloneNode(true) as SVGElement
    clone.removeAttribute('data-editor-svg')
    clone.querySelectorAll('style').forEach(s => {
      if (s.textContent?.includes('[data-selected]')) s.remove()
    })
    clone.querySelectorAll('[data-selected]').forEach(el => el.removeAttribute('data-selected'))
    const serializer = new XMLSerializer()
    return serializer.serializeToString(clone)
  }, [])

  const handleDownloadSvg = useCallback(() => {
    const str = serializeSvg()
    if (!str) return
    const blob = new Blob([str], { type: 'image/svg+xml' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `edited-${Date.now()}.svg`
    a.click()
    URL.revokeObjectURL(url)
    showToast('SVG 下载完成')
  }, [serializeSvg, showToast])

  const handleExportPng = useCallback((scale: number) => {
    const str = serializeSvg()
    if (!str) return
    const svgEl = document.querySelector('[data-editor-svg]') as SVGElement | null
    const rect = svgEl?.getBoundingClientRect()
    const w = rect?.width || 512
    const h = rect?.height || 512
    const dataUrl = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(str)
    const img = new Image()
    img.onload = () => {
      const canvas = document.createElement('canvas')
      canvas.width = w * scale
      canvas.height = h * scale
      const ctx = canvas.getContext('2d')!
      ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
      canvas.toBlob(blob => {
        if (!blob) return
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `exported-${Date.now()}.png`
        a.click()
        URL.revokeObjectURL(url)
        showToast(`PNG ${scale}x 导出完成`)
      }, 'image/png')
    }
    img.src = dataUrl
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
    if (!str) {
      setError('无法导出 SVG — 画布中无内容')
      setTimeout(() => setError(null), 3000)
      return
    }
    try {
      await svgs.save(`Edited ${new Date().toLocaleString()}`, str)
      showToast('已保存到 Library')
    } catch (err) {
      console.error('Save failed:', err)
      if (err instanceof ApiError && (err.status === 401 || err.code === 'UNAUTHORIZED')) {
        setError('登录已过期，请重新登录')
        setTimeout(() => navigate('/'), 1500)
      } else {
        setError('保存失败，请重试')
      }
      setTimeout(() => setError(null), 3000)
    }
  }, [token, navigate, showToast, serializeSvg])

  const handlePublish = useCallback(async (name: string, tags: { name: string; type: string }[], theme: string, isPublic: boolean) => {
    if (!token) { navigate('/'); return }
    const str = serializeSvg()
    if (!str) {
      setError('无法导出 SVG — 画布中无内容')
      setTimeout(() => setError(null), 3000)
      return
    }
    try {
      const res = await icons.create({ name, svg_content: str, is_public: isPublic, tags, theme })
      setShowPublish(false)
      showToast(`已发布到图标库`)
      setTimeout(() => navigate(`/icons/${res.data.id}`), 1000)
    } catch (err) {
      console.error('Publish failed:', err)
      setError('发布失败，请重试')
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
            onColorChange={(color) => {
              setCurrentColor(color)
              applyAndTick(selectedElement, color, mode)
            }}
            onAlphaChange={setAlpha}
          />
          <PresetColors onSelect={(hex) => {
            setCurrentColor(hex)
            applyAndTick(selectedElement, hex, mode)
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
            onDownloadSvg={handleDownloadSvg}
            onExportPng={handleExportPng}
            onCopy={handleCopy}
            onSave={handleSave}
            onPublish={() => setShowPublish(true)}
          />
        </SidePanel>
      </div>

      {showPublish && (
        <PublishDialog
          onClose={() => setShowPublish(false)}
          onPublish={handlePublish}
          defaultName={`Edited ${new Date().toLocaleString()}`}
        />
      )}
    </div>
  )
}
