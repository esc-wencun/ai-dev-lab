// @vitest-environment jsdom
// settings slice 单测:changeSetting / toggleTheme reducer(jsdom:reducer 内操作 documentElement/localStorage)
// localStorage('layout-setting') 预置覆盖与 isDark 存储分支的用例放独立文件 settings.storage.test.ts
// (initialState 在模块加载时读 localStorage,vitest 每文件独立模块图,避免预置值互相污染)
import { describe, expect, it } from 'vitest'
import settingsReducer, { changeSetting, toggleTheme } from './settings'
import type { SettingsState } from './settings'

function init(): SettingsState {
  return settingsReducer(undefined, { type: '@@INIT' }) as SettingsState
}

describe('settings reducer', () => {
  it('初始 state:默认配置为基准默认值;isDark 无存储方案时回落 false(jsdom 无 matchMedia,catch 分支)', () => {
    const s = init()
    expect(s.theme).toBe('#409EFF')
    expect(s.sideTheme).toBe('theme-dark')
    expect(s.navType).toBe(1)
    expect(s.tagsView).toBe(true)
    expect(s.fixedHeader).toBe(true)
    expect(s.sidebarLogo).toBe(true)
    expect(s.dynamicTitle).toBe(false)
    expect(s.isDark).toBe(false)
  })

  it('toggleTheme 翻转 isDark、切换 documentElement 的 dark class 并持久化方案', () => {
    const s0 = init()
    document.documentElement.classList.remove('dark')

    const s1 = settingsReducer(s0, toggleTheme())
    expect(s1.isDark).toBe(true)
    expect(document.documentElement.classList.contains('dark')).toBe(true)
    expect(localStorage.getItem('vueuse-color-scheme')).toBe('dark')

    const s2 = settingsReducer(s1, toggleTheme())
    expect(s2.isDark).toBe(false)
    expect(document.documentElement.classList.contains('dark')).toBe(false)
    expect(localStorage.getItem('vueuse-color-scheme')).toBe('light')
  })

  it('changeSetting 按 key 单字段更新,其余字段不受影响', () => {
    const s = init()
    const s2 = settingsReducer(s, changeSetting({ key: 'tagsView', value: false }))
    expect(s2.tagsView).toBe(false)
    // 其余字段不动
    expect(s2.theme).toBe('#409EFF')
    expect(s2.fixedHeader).toBe(true)
    expect(s2.navType).toBe(1)
  })

  it('changeSetting 连续两次各改各的字段', () => {
    let s = init()
    s = settingsReducer(s, changeSetting({ key: 'title', value: '自定义标题' }))
    s = settingsReducer(s, changeSetting({ key: 'sidebarLogo', value: false }))
    expect(s.title).toBe('自定义标题')
    expect(s.sidebarLogo).toBe(false)
    expect(s.theme).toBe('#409EFF')
  })
})
