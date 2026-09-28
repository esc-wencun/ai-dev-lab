// @vitest-environment jsdom
// settings initialState 的两个 localStorage 读取路径(layout-setting 覆盖 / vueuse-color-scheme 明暗方案)
// ——独立成文件:initialState 在模块加载时读 localStorage,vitest 每文件独立模块图,预置值不外泄
import { beforeEach, describe, expect, it, vi } from 'vitest'

function fakeMatchMedia(dark: boolean) {
  // jsdom 的 matchMedia 恒返回 matches:false,打桩固定返回,消除对 loadIsDark 回退分支的干扰
  vi.stubGlobal(
    'matchMedia',
    vi.fn().mockReturnValue({ matches: dark, addEventListener: vi.fn(), removeEventListener: vi.fn() }),
  )
}

describe('settings initialState 与 localStorage', () => {
  // initialState 在模块顶层求值:每个用例重置模块图,保证 import 重新执行 loadFromStorage/loadIsDark
  beforeEach(() => {
    vi.resetModules()
    vi.unstubAllGlobals()
    localStorage.clear()
  })
  it('layout-setting 预置值在模块加载时合并进默认配置(未覆盖项保持默认)', async () => {
    localStorage.clear()
    localStorage.setItem(
      'layout-setting',
      JSON.stringify({ theme: '#13C2C2', tagsView: false, navType: 2 }),
    )
    const mod = await import('./settings')
    const s = mod.default(undefined, { type: '@@INIT' }) as {
      theme: string
      tagsView: boolean
      navType: number
      sideTheme: string
      fixedHeader: boolean
    }
    // 覆盖项生效
    expect(s.theme).toBe('#13C2C2')
    expect(s.tagsView).toBe(false)
    expect(s.navType).toBe(2)
    // 未覆盖项保持默认
    expect(s.sideTheme).toBe('theme-dark')
    expect(s.fixedHeader).toBe(true)
  })

  it('vueuse-color-scheme = dark → isDark 初始为 true(存储方案优先于系统偏好)', async () => {
    localStorage.clear()
    localStorage.setItem('vueuse-color-scheme', 'dark')
    fakeMatchMedia(false)
    const mod = await import('./settings')
    const s = mod.default(undefined, { type: '@@INIT' }) as { isDark: boolean }
    expect(s.isDark).toBe(true)
  })

  it('vueuse-color-scheme = light → isDark 初始为 false(覆盖系统暗色偏好)', async () => {
    localStorage.clear()
    localStorage.setItem('vueuse-color-scheme', 'light')
    fakeMatchMedia(true)
    const mod = await import('./settings')
    const s = mod.default(undefined, { type: '@@INIT' }) as { isDark: boolean }
    expect(s.isDark).toBe(false)
  })

  it('非法 JSON 走防御分支(catch → {}),回落默认配置', async () => {
    localStorage.clear()
    localStorage.setItem('layout-setting', '{bad json')
    vi.resetModules()
    const mod = await import('./settings')
    const s = mod.default(undefined, { type: '@@INIT' }) as { theme: string }
    expect(s.theme).toBe('#409EFF')
  })
})
