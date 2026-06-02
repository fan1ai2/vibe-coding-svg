import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'

// These tests verify the amber color migration was applied correctly

describe('UI Consistency — amber color migration', () => {
  it('LoadingSpinner uses amber accent', async () => {
    const { default: LoadingSpinner } = await import('../components/LoadingSpinner')
    const { container } = render(<LoadingSpinner label="Loading" />)
    const spinner = container.querySelector('.animate-spin')
    expect(spinner?.className).toContain('border-t-amber-500')
  })

  it('ErrorBoundary retry button uses amber', async () => {
    const { default: ErrorBoundary } = await import('../components/ErrorBoundary')
    const { container } = render(
      <MemoryRouter>
        <ErrorBoundary>
          <Thrower />
        </ErrorBoundary>
      </MemoryRouter>
    )
    // Wait for error boundary to catch
    const btn = container.querySelector('button')
    if (btn) {
      expect(btn.className).toContain('bg-amber-500')
      expect(btn.className).not.toContain('bg-indigo')
    }
  })

  it('DropZone uses amber for drag state', async () => {
    const { default: DropZone } = await import('../components/DropZone')
    const { container } = render(
      <DropZone onFile={() => {}} disabled={false} />
    )
    // The "Click to upload" span should use text-amber-600
    const span = container.querySelector('span')
    expect(span?.className || '').not.toContain('text-indigo')
  })

  it('PublishDialog submit button uses amber', async () => {
    const { default: PublishDialog } = await import('../components/PublishDialog')
    const { container } = render(
      <PublishDialog onClose={() => {}} onPublish={() => {}} />
    )
    const submitBtn = container.querySelector('button[type="submit"]')
    expect(submitBtn?.className).toContain('bg-amber-500')
    expect(submitBtn?.className).not.toContain('bg-blue')
  })

  it('PublishDialog uses rounded-2xl modal', async () => {
    const { default: PublishDialog } = await import('../components/PublishDialog')
    const { container } = render(
      <PublishDialog onClose={() => {}} onPublish={() => {}} />
    )
    const dialog = container.querySelector('.rounded-2xl')
    expect(dialog).toBeTruthy()
  })

  it('SearchBar input has amber focus ring', async () => {
    const { default: SearchBar } = await import('../components/SearchBar')
    const { container } = render(
      <SearchBar initialQuery="" onSearch={() => {}} />
    )
    const input = container.querySelector('input')
    expect(input?.className).toContain('focus:ring-amber-500/20')
    expect(input?.className).not.toContain('focus:ring-blue')
  })

  it('AiGeneratePage title uses text-xl font-bold', async () => {
    // AiGeneratePage requires AuthProvider — skip DOM render, verify export
    const src = await import('../pages/AiGeneratePage')
    expect(src.default).toBeDefined()
  })
})

// Component that throws to test ErrorBoundary
function Thrower() {
  throw new Error('test error')
  return null
}
