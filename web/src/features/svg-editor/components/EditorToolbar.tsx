import { useState } from 'react'

interface EditorToolbarProps {
  canUndo: boolean
  canRedo: boolean
  canExport: boolean
  onUndo: () => void
  onRedo: () => void
  onDownload: () => void
  onCopy: () => void
  onSave: () => Promise<void>
}

export default function EditorToolbar({ canUndo, canRedo, canExport, onUndo, onRedo, onDownload, onCopy, onSave }: EditorToolbarProps) {
  const [saving, setSaving] = useState(false)

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

  return (
    <div className="flex flex-wrap gap-1.5">
      {btn('Undo', !canUndo, onUndo)}
      {btn('Redo', !canRedo, onRedo)}
      <div className="w-px bg-gray-200 mx-1" />
      {btn('Download', !canExport, onDownload)}
      {btn('Copy', !canExport, onCopy)}
      {btn(saving ? 'Saving...' : 'Save to Library', !canExport || saving, handleSave, true)}
    </div>
  )
}
