// permission slice —— 对位 RuoYi-Vue3 src/store/modules/permission.js
// React 无 addRoute：采用「数据就绪后全量重渲染」Gate 方案（deviations 无契约差异）。
// generateRoutes 产出三份菜单数据：sidebarRoutes / rewriteRoutes(挂载用) / topbarRoutes。
// 注意：React element 不可序列化，不放入 Redux——挂载版路由对象存模块级 Map（routeElementMap）。

import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import type { PayloadAction } from '@reduxjs/toolkit'
import { getRouters } from '@/api/menu'
import { hasPermiOr, hasRoleOr } from '@/utils/permission'
import { buildRouteObjects, setBackendMenus, type RouteObjectLite } from '@/router/gates'
import { dynamicRoutes } from '@/router/routes'

// 后端菜单原始树（component 为字符串三占位/视图路径）——侧边栏/面包屑/搜索消费
export interface MetaItem {
  title?: string
  icon?: string
  noCache?: boolean
  link?: string | null
  activeMenu?: string
  affix?: boolean
  breadcrumb?: boolean
}

export interface RouteItem {
  name?: string
  path: string
  component?: string
  hidden?: boolean
  alwaysShow?: boolean
  meta?: MetaItem
  children?: RouteItem[]
}

export interface PermissionState {
  routes: RouteItem[] // constantRoutes + 动态（面包屑/affix 提取用）
  sidebarRoutes: RouteItem[] // 侧边栏菜单树（navType 1/2 消费）
  topbarRoutes: RouteItem[] // 顶部菜单树（navType 2/3 消费）
  generated: boolean // 动态路由是否已生成（Gate 依据）
}

const initialState: PermissionState = {
  routes: [],
  sidebarRoutes: [],
  topbarRoutes: [],
  generated: false,
}

// 深拷贝（基准用 JSON 深拷贝三份，行为一致）
function deepClone<T>(data: T): T {
  return JSON.parse(JSON.stringify(data)) as T
}

// 前端自有动态路由过滤（对位 filterDynamicRoutes）：permissions → hasPermiOr，roles → hasRoleOr，都不带丢弃
export function filterDynamicRoutes(
  routes: (RouteItem & { permissions?: string[]; roles?: string[] })[],
  permissions: string[],
  roles: string[],
): RouteItem[] {
  return routes.filter((route) => {
    if (route.permissions) return hasPermiOr(permissions, route.permissions)
    if (route.roles) return hasRoleOr(roles, route.roles)
    return false
  })
}

export const generateRoutes = createAsyncThunk(
  'permission/generateRoutes',
  async ({ permissions, roles }: { permissions: string[]; roles: string[] }) => {
    const res = (await getRouters()) as unknown as { data: RouteItem[] }
    const menus = res.data || []
    // 三份深拷贝（与基准 generateRoutes 一致）
    const sidebarData = deepClone(menus)
    const rewriteData = deepClone(menus)
    const topbarData = deepClone(menus)

    // 挂载版路由对象（含 React element）进模块级变量，不进 Redux
    const routeObjects = buildRouteObjects(rewriteData)
    setBackendMenus(sidebarData)

    // 前端自有动态路由（五条 hidden 子页）按权限过滤后一并向后追加
    const allowedDynamic = filterDynamicRoutes(dynamicRoutes as never, permissions, roles)
    // element 只进模块级变量（非序列化值不能进 action payload / Redux store）
    setResolvedRoutes(routeObjects, buildRouteObjects(allowedDynamic as never))
    return {
      sidebarRoutes: sidebarData,
      topbarRoutes: topbarData,
    }
  },
)

const permissionSlice = createSlice({
  name: 'permission',
  initialState,
  reducers: {
    // navType 2 混合模式：TopNav 点一级菜单把其 children 设为侧边栏（对位 setSidebarRouters 联动）
    setSidebarRoutes(state, action: PayloadAction<RouteItem[]>) {
      state.sidebarRoutes = action.payload
    },
  },
  extraReducers: (builder) => {
    builder.addCase(generateRoutes.fulfilled, (state, action) => {
      state.sidebarRoutes = action.payload.sidebarRoutes
      state.topbarRoutes = action.payload.topbarRoutes
      state.generated = true
      // routes = 静态 + 动态（对位 setRoutes；静态表由消费方拼接）
    })
  },
})

export const { setSidebarRoutes } = permissionSlice.actions

// 供守卫/布局读取的挂载版路由对象（含 allowedDynamic）
let latestRouteObjects: RouteObjectLite[] = []
let latestAllowedDynamic: RouteObjectLite[] = []
export function setResolvedRoutes(objects: RouteObjectLite[], allowed: RouteObjectLite[]) {
  latestRouteObjects = objects
  latestAllowedDynamic = allowed
}
export function getResolvedRoutes(): { objects: RouteObjectLite[]; allowed: RouteObjectLite[] } {
  return { objects: latestRouteObjects, allowed: latestAllowedDynamic }
}

export default permissionSlice.reducer
