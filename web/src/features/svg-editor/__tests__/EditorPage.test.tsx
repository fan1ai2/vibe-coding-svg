import { describe, it, expect } from 'vitest'

describe('EditorPage (smoke)', () => {
  it('exports a default component', async () => {
    const mod = await import('../../../pages/EditorPage')
    expect(mod.default).toBeDefined()
    expect(typeof mod.default).toBe('function')
  })
})
