// 全局守卫 —— 对位 RuoYi-Vue3 src/permission.js
// 顺序严格对齐基准：setTitle → isLock 劫持 → 白名单 → getInfo+generateRoutes → 放行

import { useEffect, useRef } from 'react'
import { Navigate, useLocation } from 'react-router'
import nprogress from 'nprogress'
import 'nprogress/nprogress.css'
import { isPathMatch } from '@/utils/validate'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import { setTitle } from '@/store/modules/settings'
import { getInfo, clearUser } from '@/store/modules/user'
import { generateRoutes } from '@/store/modules/permission'

nprogress.configure({ showSpinner: false })

const whiteList = ['/login', '/register']

export default function AuthGuard({ children }: { children: React.ReactNode }) {
  const dispatch = useAppDispatch()
  const location = useLocation()
  const token = useAppSelector((s) => s.user.token)
  const roles = useAppSelector((s) => s.user.roles)
  const generated = useAppSelector((s) => s.permission.generated)
  const pathname = location.pathname

  // NProgress：导航发起瞬间 start（ done 由页面挂载时的 NProgressDone 触发，
  // 修复原先 start/done 同处一个 effect 导致进度条无人关闭、页面加载完仍慢慢爬的问题）
  const pathnameRef = useRef(pathname)
  useEffect(() => {
    if (pathnameRef.current !== pathname) {
      pathnameRef.current = pathname
      nprogress.start()
    }
  }, [pathname])

  const inWhite = whiteList.some((p) => isPathMatch(p, pathname))
  const needsInit = !!token && roles.length === 0 && !generated

  // 权限初始化（对位守卫的 getInfo + generateRoutes 时序）
  useEffect(() => {
    if (!needsInit || inWhite) return
    let cancelled = false
    ;(async () => {
      try {
        await dispatch(getInfo()).unwrap()
        const user = await import('@/store/modules/user')
        void user
        const perms = (await import('@/store')).store.getState().user
        await dispatch(generateRoutes({ permissions: perms.permissions, roles: perms.roles })).unwrap()
      } catch {
        // getInfo 失败（401 等）→ 清除本地会话再回登录（对位基准 err 分支的 await logOut()）；
        // 不清 token 的话 GuardInner 会在 /login 重定向回 / → 再次 getInfo → 死循环
        if (!cancelled) {
          dispatch(clearUser())
          window.location.href = '/login'
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [needsInit, inWhite, dispatch])

  // setTitle（对位 to.meta.title → settings.setTitle）
  useEffect(() => {
    const handle = (location.state as { handle?: { title?: string } } | null)?.handle
    void handle
    if (pathname !== '/login' && pathname !== '/register') {
      dispatch(setTitle(document.title))
    }
  }, [pathname, dispatch])

  if (!token) {
    if (inWhite) return <>{children}</>
    const redirect = encodeURIComponent(pathname + location.search)
    return <Navigate to={`/login?redirect=${redirect}`} replace />
  }
  return <GuardInner>{children}</GuardInner>
}

// 内层守卫：已登录后的分支（锁屏劫持 / 初始化 Loading / 去登录重定向 / 白名单）
function GuardInner({ children }: { children: React.ReactNode }) {
  const location = useLocation()
  const roles = useAppSelector((s) => s.user.roles)
  const generated = useAppSelector((s) => s.permission.generated)
  const isLock = useAppSelector((s) => s.lock.isLock)
  const pathname = location.pathname

  // (白名单分支仅未登录态使用，登录后统一走重定向)

  // 去 /login：已登录重定向到 /（对位基准 L26-29）
  if (pathname === '/login' || pathname === '/register') {
    return <Navigate to="/" replace />
  }
  // 锁屏硬劫持（对位基准 L33-40）
  if (isLock && pathname !== '/lock') return <Navigate to="/lock" replace />
  if (!isLock && pathname === '/lock') return <Navigate to="/" replace />
  // 初始化中：全屏 Loading（等价基准 {...to, replace:true} 重导航的等待期）
  if (roles.length === 0 && !generated) return <FullScreenLoading />
  return <>{children}</>
}

export { whiteList }

// 全屏加载态（来自 gates，单独 re-export 方便 App 使用）
import { FullScreenLoading } from './gates'

// 页面挂载即结束进度条：与懒加载页面放在同一 Suspense 边界内，
// 页面代码块加载完成、两者一起挂载 → nprogress.done() 真实反映"页面已就绪"
export function NProgressDone() {
  useEffect(() => {
    nprogress.done()
  }, [])
  return null
}
