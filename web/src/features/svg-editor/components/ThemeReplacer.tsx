import { ColorMap, ColorMode } from '../domain/types'

interface ThemeReplacerProps {
  colorMap: ColorMap
  targetColor: string
  mode: ColorMode
  onReplace: (sourceColor: string, targetColor: string, mode: ColorMode) => void
}

export default function ThemeReplacer({ colorMap, targetColor, mode, onReplace }: ThemeReplacerProps) {
  const colors = Array.from(colorMap.keys()).sort()

  const handleReplace = () => {
    const select = document.getElementById('theme-source-color') as HTMLSelectElement
    const sourceColor = select?.value
    if (sourceColor && sourceColor !== targetColor) {
      onReplace(sourceColor, targetColor, mode)
    }
  }

  if (colors.length === 0) {
    return (
      <div className="text-xs text-gray-400 text-center py-2">暂无颜色可替换</div>
    )
  }

  return (
    <div className="space-y-2">
      <label className="text-xs font-medium text-gray-500">Theme Replace</label>
      <select
        id="theme-source-color"
        className="w-full text-sm border border-gray-200 rounded-lg px-2 py-1.5 bg-white"
      >
        {colors.map(c => (
          <option key={c} value={c}>{c}</option>
        ))}
      </select>
      <button
        onClick={handleReplace}
        disabled={colors.length === 0}
        className="w-full py-1.5 text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600 disabled:bg-gray-300 rounded-lg transition-colors"
      >
        Replace All ({mode})
      </button>
    </div>
  )
}
