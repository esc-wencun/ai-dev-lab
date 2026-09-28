// useRouteMatches —— useMatches 的非 data-router 等价物
// 工程用 BrowserRouter + useRoutes(非 createBrowserRouter data router),useMatches 会抛
// "must be used within a data router";此处用 matchRoutes 对同一份全量路由表求匹配,
// 返回同形状的 { handle }[] 供 Layout/TagsView/TopNav/TopBar/useTabActions 消费。

import { useLocation, matchRoutes } from 'react-router'
import type { RouteObject } from 'react-router'
import { constantRoutes, layoutRoute, error401Route, notFoundRoute } from '@/router/routes'
import { getResolvedRoutes } from '@/store/modules/permission'
import { useAppSelector } from '@/store/hooks'
import type { RouteObjectLite } from '@/router/gates'

// 组装全量挂载版路由表(与 App.tsx 的 useRoutes 输入必须一致——App 也改为调用本函数,
// 单一来源防止组装漂移)
export function buildFullRouteObjects(generated: boolean): RouteObjectLite[] {
  const resolved = generated ? getResolvedRoutes() : { objects: [], allowed: [] }
  const mergedLayout: RouteObjectLite = {
    path: '/',
    element: layoutRoute.element,
    children: [
      ...(layoutRoute.children || []),
      ...resolved.objects,
      ...resolved.allowed,
    ],
  }
  return [...constantRoutes, mergedLayout, error401Route, notFoundRoute]
}

export interface RouteMatchLite {
  handle?: unknown
  pathname: string
  params: Record<string, string | undefined>
}

export function useRouteMatches(): RouteMatchLite[] {
  const location = useLocation()
  const generated = useAppSelector((s) => s.permission.generated)
  const routes = buildFullRouteObjects(generated)
  const matches = matchRoutes(routes as unknown as RouteObject[], location.pathname) || []
  return matches.map((m) => ({
    handle: m.route.handle,
    pathname: m.pathname,
    params: m.params,
  }))
}
