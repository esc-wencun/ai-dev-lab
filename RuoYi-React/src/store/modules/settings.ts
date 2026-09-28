// settings slice —— 对位 RuoYi-Vue3 src/store/modules/settings.js
// 默认值来自 config/defaultSettings.ts；localStorage 键 layout-setting 覆盖

import { createSlice } from '@reduxjs/toolkit'
import type { PayloadAction } from '@reduxjs/toolkit'
import type { AppSettings } from '@/config/defaultSettings'
import defaultSettings from '@/config/defaultSettings'

export interface SettingsState extends AppSettings {
  isDark: boolean
}

function loadFromStorage(): Partial<AppSettings> {
  try {
    const raw = localStorage.getItem('layout-setting')
    return raw ? JSON.parse(raw) : {}
  } catch {
    return {}
  }
}

// isDark 持久化（对位 vueuse useDark：localStorage 键 vueuse-color-scheme，缺省跟随系统偏好）
function loadIsDark(): boolean {
  try {
    const saved = localStorage.getItem('vueuse-color-scheme')
    if (saved === 'dark') return true
    if (saved === 'light') return false
  } catch {
    /* ignore */
  }
  try {
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  } catch {
    return false
  }
}

const initialState: SettingsState = {
  ...defaultSettings,
  ...loadFromStorage(),
  isDark: loadIsDark(),
}

const settingsSlice = createSlice({
  name: 'settings',
  initialState,
  reducers: {
    // 对位基准 changeSetting({key, value})：按 key 单字段更新
    changeSetting(state, action: PayloadAction<{ key: keyof SettingsState; value: SettingsState[keyof SettingsState] }>) {
      const { key, value } = action.payload
      ;(state as unknown as Record<string, unknown>)[key] = value
    },
    setTitle(state, action: PayloadAction<string>) {
      state.title = action.payload
      // dynamicTitle 行为对齐基准：开 → '页面标题 - 站点名'；关 → 站点名
      document.title = state.dynamicTitle
        ? `${action.payload} - ${import.meta.env.VITE_APP_TITLE}`
        : import.meta.env.VITE_APP_TITLE
    },
    toggleTheme(state) {
      state.isDark = !state.isDark
      document.documentElement.classList.toggle('dark', state.isDark)
      try {
        localStorage.setItem('vueuse-color-scheme', state.isDark ? 'dark' : 'light')
      } catch {
        /* ignore */
      }
    },
  },
})

export const { changeSetting, setTitle, toggleTheme } = settingsSlice.actions
export default settingsSlice.reducer
