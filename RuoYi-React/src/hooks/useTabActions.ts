// useTabActions —— 对位基准 src/plugins/tab.js（$tab 插件）
// 语义逐条对齐 §7.2：refreshPage / closePage / closeOpenPage / closeAllPage /
// closeLeftPage / closeRightPage / closeOtherPage / openPage / updatePage
// 基于现有 tagsView slice reducers 实现（不改变既有 action 名称与 payload 形状）

import { useLocation, useNavigate } from 'react-router'
import { store } from '@/store'
import { useRouteMatches } from './useRouteMatches'
import {
  addView,
  delAllViews,
  delCachedView,
  delIframeView,
  delLeftViews,
  delOthersViews,
  delRightViews,
  delView,
  updateVisitedView,
  type TagView,
} from '@/store/modules/tagsView'
import type { RouteHandle } from '@/router/gates'

// 从 useMatches 结果取叶子路由的 handle（最后一个带 handle 的 match）
export function leafHandle(matches: { handle?: unknown }[]): RouteHandle | undefined {
  for (let i = matches.length - 1; i >= 0; i--) {
    const h = matches[i].handle as RouteHandle | undefined
    if (h) return h
  }
  return undefined
}

// 当前路由 → TagView（对位 $tab 从 router.currentRoute 构造 obj）
function toQuery(search: string): Record<string, unknown> {
  return Object.fromEntries(new URLSearchParams(search))
}

export function useTabActions() {
  const navigate = useNavigate()
  const location = useLocation()
  const matches = useRouteMatches()
  const handle = leafHandle(matches)

  // 当前路由视图（name/path/fullPath/query/meta 齐备）
  const currentView = (): TagView => ({
    name: handle?.name,
    path: location.pathname,
    fullPath: location.pathname + location.search,
    title: handle?.title,
    query: toQuery(location.search),
    meta: handle?.meta as TagView['meta'],
  })

  // query 对象 → 查询串（'?' 开头；空对象返回 ''）
  const qs = (query?: Record<string, unknown>): string => {
    if (!query || Object.keys(query).length === 0) return ''
    const params = new URLSearchParams()
    for (const [k, v] of Object.entries(query)) params.set(k, String(v))
    const s = params.toString()
    return s ? '?' + s : ''
  }

  return {
    currentView,

    /** 刷新页签：delCachedView 后借 /redirect 中转页强制重挂载（对位 refreshPage） */
    refreshPage(obj?: Partial<TagView>) {
      // 防止在重定向过程中重复刷新
      if (location.pathname.startsWith('/redirect/')) return
      const view = obj ?? currentView()
      if (view.name) dispatch_delCached(view.name)
      navigate('/redirect' + (view.path || location.pathname) + qs(view.query as Record<string, unknown>), { replace: true })
    },

    /** 关闭当前页签并打开新页签（对位 closeOpenPage） */
    closeOpenPage(obj: Partial<TagView> | string) {
      dispatchDelView(currentView())
      void navigate(obj as never)
    },

    /** 关闭页签：无参 = 关当前并跳 visitedViews 末位（无则 '/'）；带参 = 只删指定（对位 closePage） */
    closePage(obj?: Partial<TagView>) {
      if (!obj) {
        dispatchDelView(currentView())
        const views = store.getState().tagsView.visitedViews
        const latest = views[views.length - 1]
        void navigate(latest?.fullPath || '/')
        return
      }
      dispatchDelView(obj)
    },

    /** 关闭所有页签（对位 closeAllPage） */
    closeAllPage() {
      store.dispatch(delAllViews())
    },

    /** 关闭左侧页签（对位 closeLeftPage） */
    closeLeftPage(obj?: Partial<TagView>) {
      store.dispatch(delLeftViews(obj || currentView()))
    },

    /** 关闭右侧页签（对位 closeRightPage） */
    closeRightPage(obj?: Partial<TagView>) {
      store.dispatch(delRightViews(obj || currentView()))
    },

    /** 关闭其他页签（对位 closeOtherPage） */
    closeOtherPage(obj?: Partial<TagView>) {
      store.dispatch(delOthersViews(obj || currentView()))
    },

    /** 打开页签（对位 openPage：addView + 跳转） */
    openPage(title: string, url: string, params?: Record<string, unknown>) {
      store.dispatch(addView({ path: url, fullPath: url, meta: { title } }))
      void navigate({ pathname: url, search: qs(params) })
    },

    /** 修改页签（对位 updatePage） */
    updatePage(obj: Partial<TagView>) {
      store.dispatch(updateVisitedView(obj))
    },
  }
}

// dispatch 包装（store 单例，避免逐个 import useAppDispatch 破坏 hook 组合性）
function dispatch_delCached(name: string) {
  store.dispatch(delCachedView(name))
}
// delView = 删 visited + 删 iframe（对位基准 delVisitedView 里同步清 iframeViews 的语义，
// 由两个既有 reducer 组合实现，不改既有 reducer 行为）
function dispatchDelView(view: Partial<TagView>) {
  store.dispatch(delView(view))
  store.dispatch(delIframeView(view))
}
