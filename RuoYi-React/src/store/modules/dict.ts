import { createSlice } from '@reduxjs/toolkit'
import type { PayloadAction } from '@reduxjs/toolkit'

// 字典缓存 slice —— 对位 RuoYi-Vue3 src/store/modules/dict.js
// 纯内存缓存（不持久化，刷新即失——与基准一致）
// 结构：{ [dictType]: DictDataOption[] }

export interface DictDataOption {
  label: string
  value: string
  elTagType?: string
  elTagClass?: string
}

export interface DictState {
  dict: Record<string, DictDataOption[]>
}

const initialState: DictState = { dict: {} }

const dictSlice = createSlice({
  name: 'dict',
  initialState,
  reducers: {
    setDict(state, action: PayloadAction<{ key: string; value: DictDataOption[] }>) {
      state.dict[action.payload.key] = action.payload.value
    },
    removeDict(state, action: PayloadAction<string>) {
      delete state.dict[action.payload]
    },
    cleanDict(state) {
      state.dict = {}
    },
  },
})

export const { setDict, removeDict, cleanDict } = dictSlice.actions

// 便捷 selector：取一个字典类型
export const selectDict = (key: string) => (state: { dict: DictState }) =>
  state.dict.dict[key]

export default dictSlice.reducer
