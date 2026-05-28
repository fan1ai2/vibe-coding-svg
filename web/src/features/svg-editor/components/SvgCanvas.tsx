import { useEffect, useRef, useCallback } from 'react'
import { parseSvg } from '../domain/svgParser'
import { COLORABLE_TAGS } from '../domain/types'

interface SvgCanvasProps {
  svg: string | null
  selectedElement: SVGElement | null
  onSelect: (el: SVGElement | null) => void
  onError: (msg: string) => void
  onSvgLoaded: (svgString: string, doc: Document) => void
}

export default function SvgCanvas({ svg, selectedElement, onSelect, onError, onSvgLoaded }: SvgCanvasProps) {
  const containerRef = useRef<HTMLDivElement>(null)

  const renderSvg = useCallback((svgString: string) => {
    try {
      const doc = parseSvg(svgString)

      const root = doc.documentElement
      if (root.hasAttribute('width') && parseFloat(root.getAttribute('width')!) > 5000) {
        root.setAttribute('width', '100%')
      }
      if (root.hasAttribute('height') && parseFloat(root.getAttribute('height')!) > 5000) {
        root.setAttribute('height', '100%')
      }

      for (const tag of COLORABLE_TAGS) {
        for (const el of doc.querySelectorAll(tag)) {
          el.addEventListener('click', (e: Event) => {
            e.stopPropagation()
            onSelect(el as SVGElement)
          })
          ;(el as SVGElement).style.cursor = 'pointer'
        }
      }

      if (containerRef.current) {
        containerRef.current.innerHTML = ''
        containerRef.current.appendChild(doc.documentElement)
        // inject selection highlight style
        const style = doc.createElement('style')
        style.textContent = '[data-selected]{outline:2px solid #3B82F6;outline-offset:1px}'
        doc.documentElement.insertBefore(style, doc.documentElement.firstChild)
      }

      onSvgLoaded(svgString, doc)
    } catch {
      onError('SVG 格式无效')
    }
  }, [onSelect, onError, onSvgLoaded])

  useEffect(() => {
    if (svg) renderSvg(svg)
  }, [svg, renderSvg])

  useEffect(() => {
    if (!containerRef.current) return
    const svgRoot = containerRef.current.querySelector('svg')
    if (!svgRoot) return
    svgRoot.querySelectorAll('[data-selected]').forEach(el => el.removeAttribute('data-selected'))
    if (selectedElement) {
      selectedElement.setAttribute('data-selected', 'true')
    }
  }, [selectedElement])

  if (!svg) {
    return (
      <div
        className="flex-1 flex items-center justify-center border-2 border-dashed border-gray-200 rounded-xl text-gray-400 text-sm min-h-[400px]"
        onDragOver={e => e.preventDefault()}
        onDrop={e => {
          e.preventDefault()
          const file = e.dataTransfer.files[0]
          if (file?.name.endsWith('.svg')) {
            const reader = new FileReader()
            reader.onload = () => {
              const text = reader.result as string
              renderSvg(text)
            }
            reader.readAsText(file)
          } else {
            onError('请拖入 .svg 文件')
          }
        }}
        onPaste={e => {
          const text = e.clipboardData?.getData('text')
          if (text?.includes('<svg')) {
            renderSvg(text)
          }
        }}
        tabIndex={0}
      >
        粘贴 SVG 代码 (Ctrl+V) 或拖拽 .svg 文件到此处
      </div>
    )
  }

  return (
    <div
      ref={containerRef}
      className="flex-1 overflow-auto bg-white rounded-xl border border-gray-200 p-4 min-h-[400px] [&_svg]:max-w-full [&_svg]:max-h-full"
      onClick={() => onSelect(null)}
    />
  )
}
