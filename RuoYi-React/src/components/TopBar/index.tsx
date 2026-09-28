// TopBar 顶部纯菜单条 —— 对位基准 layout/components/TopBar/index.vue（navType 3 纯顶部模式）
// sidebarRoutes 非 hidden 前 N 项渲染水平 Menu，溢出进「更多菜单」下拉；N = max(1, parseInt((宽/3)/85))
// 菜单项点击语义与侧边栏一致：外链 <a>，内部 Link 跳转（对位 SidebarItem 渲染规则）

import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router'
import { useRouteMatches } from '@/hooks/useRouteMatches'
import { Menu } from 'antd'
import type { ItemType } from 'antd/es/menu/interface'
import SvgIcon from '@/components/SvgIcon'
import { useAppSelector } from '@/store/hooks'
import type { RouteItem } from '@/store/modules/permission'

// 可视数量 N = max(1, parseInt((宽/3)/85))，resize 跟随（对位 setVisibleNumber）
function useVisibleNumber(): number {
  const [visibleNumber, setVisibleNumber] = useState(5)
  useEffect(() => {
    const calc = () => {
      const width = document.body.getBoundingClientRect().width / 3
      setVisibleNumber(Math.max(1, parseInt(String(width / 85))))
    }
    calc()
    window.addEventListener('resize', calc)
    return () => window.removeEventListener('resize', calc)
  }, [])
  return visibleNumber
}

export default function TopBar() {
  const sidebarRouters = useAppSelector((s) => s.permission.sidebarRoutes)
  const location = useLocation()
  const matches = useRouteMatches()
  const visibleNumber = useVisibleNumber()

  // 当前激活菜单（对位 activeMenu：meta.activeMenu 优先；取叶子路由 handle）
  const leafHandle = (() => {
    for (let i = matches.length - 1; i >= 0; i--) {
      const h = matches[i].handle as { activeMenu?: string } | undefined
      if (h) return h
    }
    return undefined
  })()
  const activeMenu = leafHandle?.activeMenu || location.pathname

  // 非 hidden 前 N 项 + 溢出（对位 topMenus / moreRoutes）
  const notHidden = sidebarRouters.filter((f) => !f.hidden)
  const visible = notHidden.slice(0, visibleNumber)
  const more = notHidden.slice(visibleNumber)

  // 菜单项构建（对位 SidebarItem：唯一可见子路由直接渲染子级；外链 <a>；有 children 折叠子菜单）
  function toMenuItem(item: RouteItem): ItemType {
    const meta = item.meta || {}
    const icon = meta.icon && meta.icon !== '#' ? <SvgIcon iconClass={meta.icon} size={14} /> : undefined
    const visibleChildren = (item.children || []).filter((c) => !c.hidden)
    const onlyOne = visibleChildren.length === 1 && !visibleChildren[0].children
    const showChildren = item.alwaysShow ? true : visibleChildren.length > 0 && !onlyOne

    if (onlyOne) {
      const child = visibleChildren[0]
      const childMeta = child.meta || {}
      const childPath = joinPath(item.path, child.path)
      const childIsHttp = /^https?:\/\//.test(child.path) || (!!childMeta.link && /^https?:\/\//.test(childMeta.link))
      const childLabel = childMeta.title || meta.title || item.path
      return {
        key: childPath,
        icon,
        label: childIsHttp ? (
          <a href={childMeta.link || child.path} target="_blank" rel="noopener">{childLabel}</a>
        ) : (
          <Link to={childPath}>{childLabel}</Link>
        ),
      }
    }

    if (showChildren) {
      return {
        key: item.path,
        icon,
        label: meta.title || item.path,
        children: visibleChildren.map((c) => {
          const childMeta = c.meta || {}
          const childPath = joinPath(item.path, c.path)
          const childIsHttp = /^https?:\/\//.test(c.path) || (!!childMeta.link && /^https?:\/\//.test(childMeta.link))
          const childLabel = childMeta.title || c.path
          return {
            key: childPath,
            label: childIsHttp ? (
              <a href={childMeta.link || c.path} target="_blank" rel="noopener">{childLabel}</a>
            ) : (
              <Link to={childPath}>{childLabel}</Link>
            ),
          }
        }),
      }
    }

    const isHttp = /^https?:\/\//.test(item.path) || (!!meta.link && /^https?:\/\//.test(meta.link))
    return {
      key: item.path,
      icon,
      label: isHttp ? (
        <a href={meta.link || item.path} target="_blank" rel="noopener">{meta.title || item.path}</a>
      ) : (
        <Link to={item.path}>{meta.title || item.path}</Link>
      ),
    }
  }

  // 拼接父子 path（对位 resolvePath 语义简化版）
  function joinPath(a: string, b: string): string {
    if (/^https?:\/\//.test(b)) return b
    return ('/' + (a + '/' + b).replace(/^\/+|\/+/g, '/')).replace(/\/$/, '') || '/'
  }

  const items: ItemType[] = [
    ...visible.map((r) => toMenuItem(r)),
    ...(more.length > 0
      ? [{ key: 'more', label: '更多菜单', children: more.map((r) => toMenuItem(r)) }]
      : []),
  ]

  return (
    <Menu
      mode="horizontal"
      items={items}
      selectedKeys={[activeMenu]}
      style={{ flex: 1, minWidth: 0, background: 'transparent', borderBottom: 'none' }}
    />
  )
}
