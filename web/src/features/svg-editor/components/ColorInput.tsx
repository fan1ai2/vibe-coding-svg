import { useState, useCallback } from 'react'
import { hexToRgb } from '../domain/colorUtils'

interface ColorInputProps {
  color: string
  onChange: (color: string) => void
}

export default function ColorInput({ color, onChange }: ColorInputProps) {
  const [input, setInput] = useState(color)
  const [error, setError] = useState(false)

  const handleChange = useCallback((value: string) => {
    setInput(value)
    const rgb = hexToRgb(value)
    if (rgb) {
      setError(false)
      onChange(value.toUpperCase())
    } else {
      setError(true)
    }
  }, [onChange])

  const rgb = hexToRgb(color)

  return (
    <div className="space-y-2">
      <div>
        <label className="text-xs font-medium text-gray-500">HEX</label>
        <input
          type="text"
          value={input}
          onChange={e => handleChange(e.target.value)}
          onBlur={() => { setInput(color); setError(false) }}
          maxLength={7}
          className={`w-full mt-0.5 px-2 py-1 text-sm border rounded font-mono ${
            error ? 'border-red-400 bg-red-50' : 'border-gray-200'
          }`}
        />
      </div>
      <div>
        <label className="text-xs font-medium text-gray-500">RGB</label>
        <div className="mt-0.5 px-2 py-1 text-sm text-gray-500 font-mono bg-gray-50 rounded border border-gray-100">
          {rgb ? `rgb(${rgb[0]}, ${rgb[1]}, ${rgb[2]})` : '—'}
        </div>
      </div>
    </div>
  )
}
