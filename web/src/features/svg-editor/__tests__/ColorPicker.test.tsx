import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import ColorPicker from '../components/ColorPicker'

describe('ColorPicker', () => {
  it('渲染默认颜色', () => {
    render(<ColorPicker color="#FF0000" alpha={100} onColorChange={() => {}} onAlphaChange={() => {}} />)
    expect(screen.getByDisplayValue('#FF0000')).toBeDefined()
  })
  it('显示 RGB 值', () => {
    render(<ColorPicker color="#FF0000" alpha={100} onColorChange={() => {}} onAlphaChange={() => {}} />)
    expect(screen.getByText('rgb(255, 0, 0)')).toBeDefined()
  })
})
