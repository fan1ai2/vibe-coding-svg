interface ColorPreviewProps {
  color: string
  alpha: number
}

function hexWithAlpha(hex: string, alpha: number): string {
  const a = Math.round((alpha / 100) * 255).toString(16).padStart(2, '0')
  return hex + a
}

export default function ColorPreview({ color, alpha }: ColorPreviewProps) {
  return (
    <div className="space-y-1">
      <label className="text-xs font-medium text-gray-500">Preview</label>
      <div className="w-16 h-16 rounded-lg overflow-hidden border border-gray-200">
        <div className="h-1/2" style={{ backgroundColor: color }} />
        <div
          className="h-1/2"
          style={{
            backgroundColor: hexWithAlpha(color, alpha),
            backgroundImage: 'repeating-conic-gradient(#ccc 0% 25%, #fff 0% 50%) 50% / 6px 6px',
            backgroundBlendMode: 'overlay',
          }}
        />
      </div>
    </div>
  )
}
