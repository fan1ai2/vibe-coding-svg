interface HueSliderProps {
  hue: number
  onChange: (hue: number) => void
}

export default function HueSlider({ hue, onChange }: HueSliderProps) {
  return (
    <div className="space-y-1">
      <label className="text-xs font-medium text-gray-500">Hue</label>
      <input
        type="range"
        min={0}
        max={360}
        value={hue}
        onChange={e => onChange(Number(e.target.value))}
        className="w-full h-3 rounded-full appearance-none cursor-pointer"
        style={{
          background: 'linear-gradient(to right, #F00, #FF0, #0F0, #0FF, #00F, #F0F, #F00)',
        }}
      />
    </div>
  )
}
