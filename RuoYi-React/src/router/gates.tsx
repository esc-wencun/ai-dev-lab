// 动态路由构建与 Gate 组件 —— 2.0.0 核心
// buildRouteObjects: 后端菜单树 → react-router 路由对象（Layout 壳 + 拍平叶子）
// DynamicRoutesGate: 未就绪渲染 Loading，就绪后全量重渲染（React 无 addRoute 的等价方案）

import { lazy, Suspense, createElement } from 'react'
import type { ReactNode } from 'react'
import { Outlet } from 'react-router'
import { Spin } from 'antd'
import { useAppSelector } from '@/store/hooks'
import type { RouteItem, MetaItem } from '@/store/modules/permission'

export interface RouteObjectLite {
  path: string
  element?: ReactNode
  children?: RouteObjectLite[]
  handle?: RouteHandle
}

export interface RouteHandle {
  title?: string
  name?: string
  meta?: MetaItem
  metaLink?: string | null
  activeMenu?: string
  noCache?: boolean
  affix?: boolean
  icon?: string
}

// ---------- 视图懒加载 ----------

const viewLoaders = import.meta.glob('../views/**/*.tsx')

export function viewKey(view: string): string {
  return `../views/${view}.tsx`
}

// component 配错或未实现时的占位页（基准 loadView 未命中 → 白屏；
// React 版给占位页，有意差异登记 deviations）。
function ViewPlaceholder() {
  return (
    <div style={{ padding: 40 }}>
      <h2>页面未找到组件映射</h2>
      <p>该菜单的 component 配置未映射到本工程实现的页面（tool/build 属有意不实现，见 deviations #2）。</p>
    </div>
  )
}

// ---------- 路由 name 派生 ----------

// 后端菜单无 name，前端按 path 派生稳定 name（tagsView/cachedViews 依赖）。
// 动态段（:userId 等）剔除后 PascalCase；'/' → 'Index'。
export function deriveRouteName(path: string): string {
  if (!path || path === '/') return 'Index'
  const stripped = path.split(':')[0]
  const segs = stripped
    .split('/')
    .filter(Boolean)
    .map((s) => s.charAt(0).toUpperCase() + s.slice(1))
  return segs.join('')
}

// ---------- 后端菜单 → react-router 路由对象 ----------

// 全局后端菜单引用（侧边栏/面包屑/搜索消费原始树）
let backendMenus: RouteItem[] = []

export function setBackendMenus(menus: RouteItem[]) {
  backendMenus = menus
}

export function getBackendMenus(): RouteItem[] {
  return backendMenus
}

// 判断菜单是否外链（meta.link 为 http/https）
function isHttpLink(meta?: MetaItem): boolean {
  const link = meta?.link
  return !!link && /^https?:\/\//.test(link)
}

// 拼接父子 path（基准 filterChildren 语义：子路径拼父路径）
function joinPath(parent: string, child: string): string {
  if (/^https?:\/\//.test(child)) return child
  if (!parent || parent === '/') return '/' + String(child).replace(/^\//, '')
  return `${parent}/${String(child).replace(/^\//, '')}`
}

// 构建单个叶子路由对象（Layout 内的页面）
function buildLeaf(
  parentPath: string,
  item: RouteItem,
): RouteObjectLite | null {
  const meta = item.meta || {}
  if (item.hidden) return null
  // 外链不注册 Route（侧边栏渲染 <a> 跳转，对齐基准 isHttp 跳过逻辑）
  if (isHttpLink(meta)) return null
  const path = joinPath(parentPath, item.path)
  const name = item.name || deriveRouteName(path)
  let element: ReactNode
  if (item.component === 'InnerLink') {
    // InnerLink：iframe 容器占位路由（IframeToggle 在 Layout 内按 meta.link 渲染）
    element = makeViewElement('error/404')
  } else if (item.component && item.component !== 'Layout' && item.component !== 'ParentView') {
    element = makeViewElement(item.component)
  } else {
    element = createElement(Outlet)
  }
  return {
    path,
    element,
    handle: {
      name,
      title: meta.title,
      meta,
      metaLink: meta.link || null,
      activeMenu: meta.activeMenu,
      noCache: meta.noCache,
      affix: meta.affix,
      icon: meta.icon,
    },
  }
}

// 拍平中间层：收集 Layout 下全部叶子（基准 type=true 的 filterChildren 等价实现）
function collectLeaves(
  parentPath: string,
  items: RouteItem[],
  out: RouteObjectLite[],
) {
  for (const item of items) {
    if (item.hidden) continue
    if (isHttpLink(item.meta)) continue
    const fullPath = joinPath(parentPath, item.path)
    if (item.children && item.children.length > 0) {
      const visible = item.children.filter((c) => !c.hidden)
      // 唯一可见子路由直接取子级路径（对齐基准 SidebarItem onlyOneChild 展示逻辑的路由侧）
      if (visible.length === 1 && !visible[0].children) {
        const leaf = buildLeaf(fullPath, visible[0])
        if (leaf) out.push(leaf)
        continue
      }
      // ParentView 中间层：递归拍平，子路径拼父路径
      collectLeaves(fullPath, item.children, out)
    } else {
      const leaf = buildLeaf(parentPath, item)
      if (leaf) out.push(leaf)
    }
  }
}

// 主入口：后端菜单树 → RouteObjectLite[]
// 结构：一个 Layout 壳路由（path '/'）挂全部叶子 + 内层已就绪的静态壳由 App 组装
export function buildRouteObjects(menus: RouteItem[]): RouteObjectLite[] {
  const leaves: RouteObjectLite[] = []
  for (const top of menus) {
    if (top.hidden) continue
    if (isHttpLink(top.meta)) continue
    const topPath = top.path
    if (top.component === 'InnerLink' && !top.children) {
      // 顶层 InnerLink（无子级）：整体作为叶子
      const leaf = buildLeaf('/', top)
      if (leaf) leaves.push(leaf)
    } else if (top.children && top.children.length > 0) {
      collectLeaves(topPath, top.children, leaves)
    } else {
      const leaf = buildLeaf('/', top)
      if (leaf) leaves.push(leaf)
    }
  }
  return leaves
}

// ---------- Gate 组件 ----------

// 动态路由 Gate：权限数据未生成时渲染全屏 Loading；就绪后由 App 重新 useRoutes 全量表。
// 本组件只负责「未就绪拦截 + 重渲染触发」（读取 generated 标志）。
export function useDynamicRoutesReady(): boolean {
  return useAppSelector((s) => s.permission.generated)
}

export function FullScreenLoading() {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh' }}>
      <Spin size="large" tip="系统加载中..." />
    </div>
  )
}

export { ViewPlaceholder }

// ---------- element 工厂 ----------

// 视图字符串 → 懒加载 element（Suspense + lazy）
export function makeViewElement(view: string): ReactNode {
  const key = viewKey(view)
  const loader = viewLoaders[key]
  if (!loader) return createElement(ViewPlaceholder)
  const Comp = lazy(loader as () => Promise<{ default: React.ComponentType }>)
  return createElement(
    Suspense,
    { fallback: createElement(Spin, { size: 'large' }) },
    createElement(Comp),
  )
}

// ---------- Layout 壳 ----------

// 布局壳（登录后页面容器）。4.0.0 将替换为完整布局（Sidebar/Navbar/TagsView），
// 当前版本提供 Outlet 出口 + 基本占位，保证 2.0.0 登录链路可走通。
export function LayoutShell() {
  return createElement('div', { className: 'app-wrapper' }, createElement(Outlet))
}
