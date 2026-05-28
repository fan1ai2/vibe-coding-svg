import { useState, useCallback } from 'react'
import { hsvToRgb, rgbToHex, rgbToHsv, hexToRgb } from '../domain/colorUtils'
import HueSlider from './HueSlider'
import SBPanel from './SBPanel'
import AlphaSlider from './AlphaSlider'
import ColorPreview from './ColorPreview'
import ColorInput from './ColorInput'

interface ColorPickerProps {
  color: string
  alpha: number
  onColorChange: (color: string) => void
  onAlphaChange: (alpha: number) => void
}

export default function ColorPicker({ color, alpha, onColorChange, onAlphaChange }: ColorPickerProps) {
  const rgb = hexToRgb(color) ?? [255, 0, 0]
  const [h, s, v] = rgbToHsv(rgb[0], rgb[1], rgb[2])

  const [hue, setHue] = useState(h)
  const [sat, setSat] = useState(s)
  const [bri, setBri] = useState(v)

  const handleSBChange = useCallback((newSat: number, newBri: number) => {
    setSat(newSat)
    setBri(newBri)
    const [r, g, b] = hsvToRgb(hue, newSat, newBri)
    onColorChange(rgbToHex(r, g, b))
  }, [hue, onColorChange])

  const handleHueChange = useCallback((newHue: number) => {
    setHue(newHue)
    const [r, g, b] = hsvToRgb(newHue, sat, bri)
    onColorChange(rgbToHex(r, g, b))
  }, [sat, bri, onColorChange])

  return (
    <div className="space-y-4">
      <div className="flex gap-4">
        <div className="flex-1 space-y-3">
          <HueSlider hue={hue} onChange={handleHueChange} />
          <SBPanel hue={hue} saturation={sat} brightness={bri} onChange={handleSBChange} />
          <AlphaSlider alpha={alpha} color={color} onChange={onAlphaChange} />
        </div>
        <ColorPreview color={color} alpha={alpha} />
      </div>
      <ColorInput color={color} onChange={onColorChange} />
    </div>
  )
}
