import { PRESETS } from '../domain/types'

interface PresetColorsProps {
  onSelect: (hex: string) => void
}

export default function PresetColors({ onSelect }: PresetColorsProps) {
  return (
    <div className="space-y-1">
      <label className="text-xs font-medium text-gray-500">Presets</label>
      <div className="grid grid-cols-4 gap-1.5">
        {PRESETS.map(p => (
          <button
            key={p.hex}
            onClick={() => onSelect(p.hex)}
            title={p.name}
            className="w-full aspect-square rounded-md border border-gray-200 hover:scale-110 transition-transform cursor-pointer"
            style={{ backgroundColor: p.hex }}
          />
        ))}
      </div>
    </div>
  )
}
