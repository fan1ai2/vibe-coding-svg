import { useState, useRef, useEffect } from 'react'

interface EditorToolbarProps {
  canUndo: boolean
  canRedo: boolean
  canExport: boolean
  onUndo: () => void
  onRedo: () => void
  onDownloadSvg: () => void
  onExportPng: (scale: number) => void
  onCopy: () => void
  onSave: () => Promise<void>
  onPublish: () => void
}

export default function EditorToolbar({ canUndo, canRedo, canExport, onUndo, onRedo, onDownloadSvg, onExportPng, onCopy, onSave, onPublish }: EditorToolbarProps) {
  const [saving, setSaving] = useState(false)
  const [downloadOpen, setDownloadOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDownloadOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const handleSave = async () => {
    setSaving(true)
    try { await onSave() } finally { setSaving(false) }
  }

  const btn = (label: string, disabled: boolean, onClick: () => void, primary = false) => (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed ${
        primary
          ? 'bg-amber-500 text-white hover:bg-amber-600'
          : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
      }`}
    >
      {label}
    </button>
  )

  const dropItem = (label: string, onClick: () => void) => (
    <button
      onClick={() => { onClick(); setDownloadOpen(false) }}
      className="w-full text-left px-3 py-1.5 text-xs text-gray-600 hover:bg-gray-100 rounded transition-colors"
    >
      {label}
    </button>
  )

  return (
    <div className="flex flex-wrap gap-1.5">
      {btn('撤销', !canUndo, onUndo)}
      {btn('重做', !canRedo, onRedo)}
      <div className="w-px bg-gray-200 mx-1" />
      <div className="relative" ref={dropdownRef}>
        <button
          onClick={() => setDownloadOpen(!downloadOpen)}
          disabled={!canExport}
          className="px-3 py-1.5 text-xs font-medium rounded-lg bg-gray-100 text-gray-600 hover:bg-gray-200 transition-colors disabled:opacity-40 disabled:cursor-not-allowed inline-flex items-center gap-1"
        >
          下载
          <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        {downloadOpen && (
          <div className="absolute bottom-full mb-1 left-0 bg-white rounded-lg shadow-lg border border-gray-200 py-1 min-w-[120px] z-10">
            {dropItem('SVG 源文件', onDownloadSvg)}
            <div className="border-t border-gray-100 my-0.5" />
            {dropItem('PNG (1x)', () => onExportPng(1))}
            {dropItem('PNG (2x)', () => onExportPng(2))}
            {dropItem('PNG (4x)', () => onExportPng(4))}
          </div>
        )}
      </div>
      {btn('复制', !canExport, onCopy)}
      {btn(saving ? '保存中...' : '保存到 Library', !canExport || saving, handleSave, true)}
      <div className="w-px bg-gray-200 mx-1" />
      {btn('发布到图标库', !canExport, onPublish)}
    </div>
  )
}
