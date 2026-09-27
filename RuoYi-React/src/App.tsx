// App 组装 —— useRoutes 全量路由表（Gate 方案核心：数据就绪后重新求值路由表）
// constantRoutes + layoutRoute + 动态菜单叶子（getResolvedRoutes 模块级读取）+ 404 置底

import { useRoutes } from 'react-router'
import AuthGuard from './router/AuthGuard'
import { constantRoutes, layoutRoute, error401Route, notFoundRoute } from './router/routes'
import { getResolvedRoutes } from '@/store/modules/permission'
import { useAppSelector } from './store/hooks'
import type { RouteObjectLite } from './router/gates'

export default function App() {
  const generated = useAppSelector((s) => s.permission.generated)
  const menuVersion = useAppSelector((s) => s.permission.topbarRoutes.length)

  // 挂载版路由表：generated 翻转 / 菜单变化时重新求值——等价 addRoute 后的重渲染
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
  const routeObjects: RouteObjectLite[] = [
    ...constantRoutes,
    mergedLayout,
    error401Route,
    notFoundRoute,
  ]
  void menuVersion

  return (
    <AuthGuard>
      <RoutesElement routes={routeObjects} />
    </AuthGuard>
  )
}

// useRoutes 包装组件
function RoutesElement({ routes }: { routes: RouteObjectLite[] }) {
  return useRoutes(routes as never)
}
