// lock slice —— 对位 RuoYi-Vue3 src/store/modules/lock.js
// localStorage 键 screen-lock / screen-lock-path（键名是契约）

import { createSlice } from '@reduxjs/toolkit'
import type { PayloadAction } from '@reduxjs/toolkit'

export interface LockState {
  isLock: boolean
  lockPath: string
}

function loadLock(): LockState {
  try {
    const isLock = JSON.parse(localStorage.getItem('screen-lock') || 'false') as boolean
    const lockPath = localStorage.getItem('screen-lock-path') || '/index'
    return { isLock: isLock === true, lockPath }
  } catch {
    return { isLock: false, lockPath: '/index' }
  }
}

const initialState: LockState = loadLock()

const lockSlice = createSlice({
  name: 'lock',
  initialState,
  reducers: {
    // 锁屏：记录当前路径（解锁后回跳）
    lockScreen(state, action: PayloadAction<string>) {
      state.isLock = true
      state.lockPath = action.payload || '/index'
      localStorage.setItem('screen-lock', 'true')
      localStorage.setItem('screen-lock-path', state.lockPath)
    },
    // 解锁（登录成功也会调用——与基准 login 后 unlockScreen 行为一致）
    unlockScreen(state) {
      state.isLock = false
      state.lockPath = '/index'
      localStorage.setItem('screen-lock', 'false')
      localStorage.setItem('screen-lock-path', '/index')
    },
  },
})

export const { lockScreen, unlockScreen } = lockSlice.actions
export default lockSlice.reducer
