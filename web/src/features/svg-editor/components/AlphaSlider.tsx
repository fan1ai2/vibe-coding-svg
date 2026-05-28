interface AlphaSliderProps {
  alpha: number
  color: string
  onChange: (alpha: number) => void
}

export default function AlphaSlider({ alpha, color, onChange }: AlphaSliderProps) {
  return (
    <div className="space-y-1">
      <label className="text-xs font-medium text-gray-500">Alpha ({alpha}%)</label>
      <div
        className="w-full h-3 rounded-full relative cursor-pointer"
        style={{
          background: `linear-gradient(to right, transparent, ${color}),
            repeating-conic-gradient(#ccc 0% 25%, #fff 0% 50%) 50% / 8px 8px`,
        }}
      >
        <input
          type="range"
          min={0}
          max={100}
          value={alpha}
          onChange={e => onChange(Number(e.target.value))}
          className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
        />
      </div>
    </div>
  )
}
