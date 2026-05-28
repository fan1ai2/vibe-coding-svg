import { useRef, useCallback, useEffect } from 'react'

interface SBPanelProps {
  hue: number
  saturation: number
  brightness: number
  onChange: (saturation: number, brightness: number) => void
}

const PANEL_SIZE = 200
const HANDLE_SIZE = 12

export default function SBPanel({ hue, saturation, brightness, onChange }: SBPanelProps) {
  const panelRef = useRef<HTMLDivElement>(null)
  const dragging = useRef(false)

  const updateFromMouse = useCallback((clientX: number, clientY: number) => {
    const panel = panelRef.current
    if (!panel) return
    const rect = panel.getBoundingClientRect()
    const x = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width))
    const y = Math.max(0, Math.min(1, (clientY - rect.top) / rect.height))
    onChange(Math.round(x * 100), Math.round((1 - y) * 100))
  }, [onChange])

  const onMouseDown = useCallback((e: React.MouseEvent) => {
    dragging.current = true
    updateFromMouse(e.clientX, e.clientY)
  }, [updateFromMouse])

  useEffect(() => {
    const onMove = (e: MouseEvent) => {
      if (!dragging.current) return
      updateFromMouse(e.clientX, e.clientY)
    }
    const onUp = () => { dragging.current = false }
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
    return () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }
  }, [updateFromMouse])

  const sx = (saturation / 100) * PANEL_SIZE
  const sy = ((100 - brightness) / 100) * PANEL_SIZE

  return (
    <div className="space-y-1">
      <label className="text-xs font-medium text-gray-500">Saturation × Brightness</label>
      <div
        ref={panelRef}
        onMouseDown={onMouseDown}
        className="relative rounded cursor-crosshair select-none"
        style={{
          width: PANEL_SIZE,
          height: PANEL_SIZE,
          background: `linear-gradient(to top, #000, transparent),
            linear-gradient(to right, #fff, hsl(${hue}, 100%, 50%))`,
        }}
      >
        <div
          className="absolute rounded-full border-2 border-white shadow-md pointer-events-none"
          style={{
            width: HANDLE_SIZE,
            height: HANDLE_SIZE,
            left: sx - HANDLE_SIZE / 2,
            top: sy - HANDLE_SIZE / 2,
            background: `hsl(${hue}, ${saturation}%, ${brightness}%)`,
          }}
        />
      </div>
    </div>
  )
}
