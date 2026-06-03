import { describe, it, expect } from 'vitest'

describe('EditorPage（冒烟测试）', () => {
  it('导出默认组件', async () => {
    const mod = await import('../../../pages/EditorPage')
    expect(mod.default).toBeDefined()
    expect(typeof mod.default).toBe('function')
  })
})
