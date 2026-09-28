// App 组装 —— useRoutes 全量路由表（Gate 方案核心：数据就绪后重新求值路由表）
// constantRoutes + layoutRoute + 动态菜单叶子（getResolvedRoutes 模块级读取）+ 404 置底

import { useRoutes } from 'react-router'
import AuthGuard from './router/AuthGuard'
import { useAppSelector } from './store/hooks'
import { buildFullRouteObjects } from '@/hooks/useRouteMatches'

export default function App() {
  const generated = useAppSelector((s) => s.permission.generated)
  const menuVersion = useAppSelector((s) => s.permission.topbarRoutes.length)

  // 挂载版路由表：generated 翻转 / 菜单变化时重新求值——等价 addRoute 后的重渲染
  // 组装逻辑在 buildFullRouteObjects(useRouteMatches 的 matchRoutes 用同一份,保证两处一致)
  const routeObjects = buildFullRouteObjects(generated)
  void menuVersion

  return (
    <AuthGuard>
      <RoutesElement routes={routeObjects} />
    </AuthGuard>
  )
}

// useRoutes 包装组件
function RoutesElement({ routes }: { routes: Parameters<typeof useRoutes>[0] }) {
  return useRoutes(routes)
}
