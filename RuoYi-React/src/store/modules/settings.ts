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

const initialState: SettingsState = {
  ...defaultSettings,
  ...loadFromStorage(),
  isDark: document.documentElement.classList.contains('dark'),
}

const settingsSlice = createSlice({
  name: 'settings',
  initialState,
  reducers: {
    changeSetting(state, action: PayloadAction<{ key: keyof SettingsState; value: SettingsState[keyof SettingsState] }>) {
      Object.assign(state, action.payload)
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
    },
  },
})

export const { changeSetting, setTitle, toggleTheme } = settingsSlice.actions
export default settingsSlice.reducer
