import { describe, expect, it } from 'vitest'

// 冒烟测试：验证 Vitest 可运行（占位断言）
describe('smoke', () => {
  it('vitest works', () => {
    expect(1 + 1).toBe(2)
    expect(true).toBe(true)
  })
})
