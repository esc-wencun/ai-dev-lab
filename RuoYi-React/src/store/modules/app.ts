// app slice —— 对位 RuoYi-Vue3 src/store/modules/app.js（行为一致）
// Cookie 键 sidebarStatus / size 与基准一致；基准 Pinia action 亦直接写 Cookie，
// 此处为行为对齐在 reducer 内写入（刻意偏离 RTK 纯度建议，非疏忽）。

import { createSlice } from '@reduxjs/toolkit'
import type { PayloadAction } from '@reduxjs/toolkit'
import Cookies from 'js-cookie'

export interface SidebarState {
  opened: boolean
  withoutAnimation: boolean
  hide: boolean
}

export interface AppState {
  sidebar: SidebarState
  device: 'desktop' | 'mobile'
  size: string
}

const initialState: AppState = {
  sidebar: {
    opened: Cookies.get('sidebarStatus') !== '0',
    withoutAnimation: false,
    hide: false,
  },
  device: 'desktop',
  size: Cookies.get('size') || 'default',
}

const appSlice = createSlice({
  name: 'app',
  initialState,
  reducers: {
    toggleSideBar(state) {
      state.sidebar.opened = !state.sidebar.opened
      state.sidebar.withoutAnimation = false
      Cookies.set('sidebarStatus', state.sidebar.opened ? '1' : '0')
    },
    closeSideBar(state, action: PayloadAction<{ withoutAnimation?: boolean }>) {
      state.sidebar.opened = false
      state.sidebar.withoutAnimation = !!action.payload?.withoutAnimation
      Cookies.set('sidebarStatus', '0')
    },
    toggleDevice(state, action: PayloadAction<'desktop' | 'mobile'>) {
      state.device = action.payload
    },
    setSize(state, action: PayloadAction<string>) {
      state.size = action.payload
      Cookies.set('size', action.payload)
    },
    toggleSideBarHide(state, action: PayloadAction<boolean>) {
      state.sidebar.hide = action.payload
    },
  },
})

export const { toggleSideBar, closeSideBar, toggleDevice, setSize, toggleSideBarHide } =
  appSlice.actions
export default appSlice.reducer
