import { ColorMode } from '../domain/types'

interface FillStrokeTabsProps {
  mode: ColorMode
  onChange: (mode: ColorMode) => void
}

export default function FillStrokeTabs({ mode, onChange }: FillStrokeTabsProps) {
  const base = 'flex-1 px-3 py-1.5 text-sm font-medium rounded-lg transition-colors'
  const active = 'bg-amber-100 text-amber-700'
  const inactive = 'text-gray-500 hover:bg-gray-100'

  return (
    <div className="flex gap-1 bg-gray-50 rounded-lg p-0.5">
      <button
        onClick={() => onChange('fill')}
        className={`${base} ${mode === 'fill' ? active : inactive}`}
      >
        Fill
      </button>
      <button
        onClick={() => onChange('stroke')}
        className={`${base} ${mode === 'stroke' ? active : inactive}`}
      >
        Stroke
      </button>
    </div>
  )
}
