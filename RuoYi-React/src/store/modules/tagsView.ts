// tagsView slice —— 对位 RuoYi-Vue3 src/store/modules/tagsView.js
// visitedViews/cachedViews(name)/iframeViews；持久化受 settings.tagsViewPersist 控制

import { createSlice } from '@reduxjs/toolkit'
import type { PayloadAction } from '@reduxjs/toolkit'

export interface TagView {
  name?: string
  path: string
  fullPath?: string
  title?: string
  query?: Record<string, unknown>
  meta?: {
    affix?: boolean
    noCache?: boolean
    link?: string | null
    [k: string]: unknown
  }
}

export interface TagsViewState {
  visitedViews: TagView[]
  cachedViews: string[]
  iframeViews: TagView[]
}

const state: TagsViewState = {
  visitedViews: [],
  cachedViews: [],
  iframeViews: [],
}

const initialState: TagsViewState = state

// 判重按 path；title 缺省 'no-name'（与基准一致）
function findView(views: TagView[], path: string): TagView | undefined {
  return views.find((v) => v.path === path)
}

function toTagView(route: Partial<TagView>): TagView {
  return {
    name: route.name,
    path: route.path || '/',
    fullPath: route.fullPath || route.path || '/',
    title: route.title || 'no-name',
    query: route.query,
    meta: route.meta,
  }
}

export const tagsViewSlice = createSlice({
  name: 'tags-view',
  initialState,
  reducers: {
    addView(state, action: PayloadAction<Partial<TagView>>) {
      const view = toTagView(action.payload)
      if (!findView(state.visitedViews, view.path)) {
        state.visitedViews.push(view)
      }
      // cachedViews：!meta.noCache 才缓存组件名（KeepAlive 暂缓，数据面照常维护）
      if (view.name && !view.meta?.noCache && !state.cachedViews.includes(view.name)) {
        state.cachedViews.push(view.name)
      }
      if (view.meta?.link && !findView(state.iframeViews, view.path)) {
        state.iframeViews.push(view)
      }
    },
    addAffixView(state, action: PayloadAction<Partial<TagView>>) {
      const view = toTagView(action.payload)
      if (!findView(state.visitedViews, view.path)) {
        state.visitedViews.unshift(view)
      }
    },
    delView(state, action: PayloadAction<Partial<TagView>>) {
      const path = action.payload.path || '/'
      const i = state.visitedViews.findIndex((v) => v.path === path)
      if (i > -1) state.visitedViews.splice(i, 1)
      // cachedViews 清理由调用方按需触发（delVisitedView/delCachedView），与基准 tab.js 流程一致
    },
    delVisitedView(state, action: PayloadAction<string>) {
      const i = state.visitedViews.findIndex((v) => v.path === action.payload)
      if (i > -1) state.visitedViews.splice(i, 1)
    },
    delCachedView(state, action: PayloadAction<string | undefined>) {
      if (!action.payload) return
      const i = state.cachedViews.indexOf(action.payload)
      if (i > -1) state.cachedViews.splice(i, 1)
    },
    // 关闭其他：保留 affix + 当前；iframeViews 只留当前
    delOthersViews(state, action: PayloadAction<Partial<TagView>>) {
      const view = toTagView(action.payload)
      state.visitedViews = state.visitedViews.filter(
        (v) => v.meta?.affix || v.path === view.path,
      )
      state.iframeViews = state.iframeViews.filter((v) => v.path === view.path)
      state.cachedViews = state.cachedViews.filter((name) =>
        state.visitedViews.some((v) => v.name === name),
      )
    },
    // 全部关闭：visited 只留 affix；iframeViews 清空
    delAllViews(state) {
      state.visitedViews = state.visitedViews.filter((v) => v.meta?.affix)
      state.iframeViews = []
      state.cachedViews = state.cachedViews.filter((name) =>
        state.visitedViews.some((v) => v.name === name),
      )
    },
    delLeftViews(state, action: PayloadAction<Partial<TagView>>) {
      const view = toTagView(action.payload)
      const idx = state.visitedViews.findIndex((v) => v.path === view.path)
      state.visitedViews = state.visitedViews.filter(
        (v, i) => v.meta?.affix || i >= idx,
      )
      syncCached(state)
    },
    delRightViews(state, action: PayloadAction<Partial<TagView>>) {
      const view = toTagView(action.payload)
      const idx = state.visitedViews.findIndex((v) => v.path === view.path)
      state.visitedViews = state.visitedViews.filter(
        (v, i) => v.meta?.affix || i <= idx,
      )
      syncCached(state)
    },
    updateVisitedView(state, action: PayloadAction<Partial<TagView>>) {
      const v = toTagView(action.payload)
      const target = findView(state.visitedViews, v.path)
      if (target) Object.assign(target, v)
    },
    delIframeView(state, action: PayloadAction<Partial<TagView>>) {
      const view = toTagView(action.payload)
      state.iframeViews = state.iframeViews.filter((v) => v.path !== view.path)
    },
    // 恢复持久化页签（对位基准 loadPersistedViews：按 path 判重逐条追加，affix 由调用方随后 unshift）
    loadPersistedViews(state, action: PayloadAction<TagView[]>) {
      for (const view of action.payload) {
        if (!findView(state.visitedViews, view.path)) {
          state.visitedViews.push(toTagView(view))
        }
      }
    },
  },
})

// cachedViews 与现存 visitedViews 同步（关页签后清孤儿缓存名）
function syncCached(state: TagsViewState) {
  state.cachedViews = state.cachedViews.filter((name) =>
    state.visitedViews.some((v) => v.name === name),
  )
}

export const {
  addView,
  addAffixView,
  delView,
  delVisitedView,
  delCachedView,
  delOthersViews,
  delAllViews,
  delLeftViews,
  delRightViews,
  updateVisitedView,
  delIframeView,
  loadPersistedViews,
} = tagsViewSlice.actions

export default tagsViewSlice.reducer
