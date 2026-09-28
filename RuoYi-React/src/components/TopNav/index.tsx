// TopNav 顶部横向一级菜单 —— 对位基准 layout/components/TopNav/index.vue（navType 2 混合模式）
// 数据源 permission.topbarRoutes 非 hidden 项；path==='/' 取 children[0]；
// 点有 children 的一级菜单 → setSidebarRoutes(children 加工后) + hide=false；
// 无 children 内部打开 + hide=true；http(s) 外链 window.open；溢出项进「更多菜单」；
// visibleNumber = max(1, parseInt((宽/3)/85))，resize 跟随

import { useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate } from 'react-router'
import { useRouteMatches } from '@/hooks/useRouteMatches'
import { Menu } from 'antd'
import type { ItemType } from 'antd/es/menu/interface'
import SvgIcon from '@/components/SvgIcon'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import { setSidebarRoutes } from '@/store/modules/permission'
import { toggleSideBarHide } from '@/store/modules/app'
import type { RouteItem } from '@/store/modules/permission'

// 隐藏侧边栏路由（不参与一级联动/高亮拆分，对位基准 hideList）
const hideList = ['/index', '/user/profile']

type ChildItem = RouteItem & { parentPath?: string }

export default function TopNav() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const location = useLocation()
  const matches = useRouteMatches()
  const routers = useAppSelector((s) => s.permission.topbarRoutes)

  // 顶部可视数量（对位 setVisibleNumber）
  const [visibleNumber, setVisibleNumber] = useState(1)

  useEffect(() => {
    const calc = () => {
      const width = document.body.getBoundingClientRect().width / 3
      setVisibleNumber(Math.max(1, parseInt(String(width / 85))))
    }
    calc()
    window.addEventListener('resize', calc)
    return () => window.removeEventListener('resize', calc)
  }, [])

  // 顶部菜单（对位 topMenus：非 hidden；path==='/' 取 children[0]）
  const topMenus = useMemo<RouteItem[]>(() => {
    const out: RouteItem[] = []
    for (const menu of routers) {
      if (menu.hidden === true) continue
      if (menu.path === '/' && menu.children && menu.children.length > 0) {
        out.push(menu.children[0])
      } else {
        out.push(menu)
      }
    }
    return out
  }, [routers])

  // 全部子路由展开（对位 childrenMenus：父路径拼接 + parentPath 标记）
  // 注意：topbarRoutes 是 RTK 深冻结 state，必须先克隆再加工（基准 Pinia 可直接改，此处不能）
  const childrenMenus = useMemo<ChildItem[]>(() => {
    const out: ChildItem[] = []
    for (const router of routers) {
      for (const child of router.children || []) {
        const item = JSON.parse(JSON.stringify(child)) as ChildItem
        if (item.parentPath === undefined) {
          if (router.path === '/') {
            item.path = '/' + String(item.path).replace(/^\//, '')
          } else if (!/^https?:\/\//.test(item.path)) {
            item.path = router.path + '/' + String(item.path).replace(/^\//, '')
          }
          item.parentPath = router.path
        }
        out.push(item)
      }
    }
    return out
  }, [routers])

  // 当前路由是否 InnerLink（对位 !route.meta.link 判断；叶子 handle.metaLink）
  const isMetaLink = useMemo(() => {
    for (let i = matches.length - 1; i >= 0; i--) {
      const h = matches[i].handle as { metaLink?: string | null } | undefined
      if (h) return !!h.metaLink
    }
    return false
  }, [matches])

  // 默认激活菜单（对位 activeMenu computed 的纯计算部分）
  const activeMenu = useMemo(() => {
    const path = location.pathname
    let activePath = path
    if (path && path.lastIndexOf('/') > 0 && !hideList.includes(path)) {
      const tmpPath = path.substring(1)
      if (!isMetaLink) {
        activePath = '/' + tmpPath.substring(0, tmpPath.indexOf('/'))
      }
    }
    return activePath
  }, [location.pathname, isMetaLink])

  // activeMenu computed 的副作用部分（对位 appStore.toggleSideBarHide 联动）
  useEffect(() => {
    const path = location.pathname
    if (path && path.lastIndexOf('/') > 0 && !hideList.includes(path)) {
      if (!isMetaLink) dispatch(toggleSideBarHide(false))
    } else {
      dispatch(toggleSideBarHide(true))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.pathname, isMetaLink])

  // 联动侧边栏（对位 activeRoutes(key)）
  const activeRoutes = (key: string): RouteItem[] => {
    const routes: RouteItem[] = []
    for (const item of childrenMenus) {
      if (key === item.parentPath || (key === 'index' && item.path === '')) {
        routes.push(item)
      }
    }
    if (routes.length > 0) {
      dispatch(setSidebarRoutes(routes))
    } else {
      dispatch(toggleSideBarHide(true))
    }
    return routes
  }

  // 点击一级菜单（对位 handleSelect）
  const handleSelect = (key: string) => {
    const route = routers.find((item) => item.path === key)
    if (/^https?:\/\//.test(key)) {
      // http(s) 路径新窗口打开
      window.open(key, '_blank')
      return
    }
    if (!route || !route.children) {
      // 没有子路由路径内部打开
      const routeMenu = childrenMenus.find((item) => item.path === key)
      const query = routeMenu?.meta && (routeMenu.meta as { query?: string }).query
      if (routeMenu && query) {
        try {
          void navigate({ pathname: key, search: new URLSearchParams(JSON.parse(query as string)).toString() ? toSearch(JSON.parse(query as string)) : '' })
        } catch {
          void navigate(key)
        }
      } else {
        void navigate(key)
      }
      dispatch(toggleSideBarHide(true))
    } else {
      // 显示左侧联动菜单
      activeRoutes(key)
      dispatch(toggleSideBarHide(false))
    }
  }

  // query 对象 → search 串
  const toSearch = (query: Record<string, unknown>): string => {
    const params = new URLSearchParams()
    for (const [k, v] of Object.entries(query)) params.set(k, String(v))
    const s = params.toString()
    return s ? '?' + s : ''
  }

  // 菜单项构建（先于 items 声明，避免 TDZ）
  const toMenuItem = (menu: RouteItem): ItemType => {
    const meta = menu.meta || {}
    const icon = meta.icon && meta.icon !== '#' ? <SvgIcon iconClass={meta.icon} size={14} /> : undefined
    return { key: menu.path, icon, label: meta.title || menu.path }
  }

  // antd Menu items：前 N 项 + 溢出折叠进「更多菜单」
  const items: ItemType[] = [
    ...topMenus.slice(0, visibleNumber).map((menu) => toMenuItem(menu)),
    ...(topMenus.length > visibleNumber
      ? [
          {
            key: 'more',
            label: '更多菜单',
            children: topMenus.slice(visibleNumber).map((menu) => toMenuItem(menu)),
          },
        ]
      : []),
  ]

  return (
    <Menu
      mode="horizontal"
      items={items}
      selectedKeys={[activeMenu]}
      style={{ flex: 1, minWidth: 0, borderBottom: 'none', background: 'transparent' }}
      onClick={({ key }) => handleSelect(String(key))}
    />
  )
}
