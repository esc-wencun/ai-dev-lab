// Sidebar —— 对位基准 layout/components/Sidebar + SidebarItem
// 菜单数据来自 permission store（后端菜单树）；唯一可见子路由直接渲染子级；外链 <a>

import { Menu } from 'antd'
import type { ItemType } from 'antd/es/menu/interface'
import { Link, useNavigate } from 'react-router'
import SvgIcon from '@/components/SvgIcon'
import { useAppSelector } from '@/store/hooks'
import type { RouteItem } from '@/store/modules/permission'
import type { MetaItem } from '@/store/modules/permission'

interface SidebarProps {
  collapsed: boolean
  currentPath: string
}

// 递归构建 antd Menu items（对位 SidebarItem 递归渲染）
function buildItems(items: RouteItem[], parentPath: string): ItemType[] {
  return items
    .filter((r) => !r.hidden)
    .map((r) => {
      const meta: MetaItem = r.meta || {}
      const fullPath = /^https?:\/\//.test(r.path)
        ? r.path
        : ('/' + (parentPath + '/' + r.path).replace(/^\/+|\/+/g, '/').replace(/^\//, ''))
      const icon = meta.icon ? <SvgIcon iconClass={meta.icon} size={16} /> : undefined
      const label = meta.title || r.path
      const visibleChildren = (r.children || []).filter((c) => !c.hidden)
      const onlyOne = visibleChildren.length === 1 && !visibleChildren[0].children
      const showChildren = r.alwaysShow ? true : visibleChildren.length > 0 && !onlyOne

      if (onlyOne) {
        const child = visibleChildren[0]
        const childMeta = child.meta || {}
        const childPath = /^https?:\/\//.test(child.path)
          ? child.path
          : joinPath(fullPath, child.path)
        const childIsHttp = /^https?:\/\//.test(child.path) || (!!childMeta.link && /^https?:\/\//.test(childMeta.link))
        return {
          key: childPath,
          icon,
          label: childIsHttp ? (
            <a href={childMeta.link || child.path} target="_blank" rel="noopener">{childMeta.title || label}</a>
          ) : (
            <Link to={childPath}>{childMeta.title || childMeta.title || label}</Link>
          ),
        }
      }

      if (showChildren) {
        return {
          key: fullPath,
          icon,
          label,
          children: buildItems(r.children || [], fullPath),
        }
      }

      const isHttp = /^https?:\/\//.test(r.path) || (!!meta.link && /^https?:\/\//.test(meta.link))
      return {
        key: fullPath,
        icon,
        label: isHttp ? (
          <a href={meta.link || r.path} target="_blank" rel="noopener">{label}</a>
        ) : (
          <Link to={fullPath}>{label}</Link>
        ),
      }
    })
}

function joinPath(a: string, b: string): string {
  if (/^https?:\/\//.test(b)) return b
  return ('/' + (a + '/' + b).replace(/^\/+|\/+/g, '/')).replace(/\/$/, '') || '/'
}

export default function Sidebar({ collapsed, currentPath }: SidebarProps) {
  const navigate = useNavigate()
  const sidebarRoutes = useAppSelector((s) => s.permission.sidebarRoutes)
  const sideTheme = useAppSelector((s) => s.settings.sideTheme)
  const isDark = useAppSelector((s) => s.settings.isDark)
  const sidebarLogo = useAppSelector((s) => s.settings.sidebarLogo)

  const items = buildItems(sidebarRoutes, '')

  return (
    <div
      style={{
        width: collapsed ? 54 : 200,
        transition: 'width 0.2s',
        background: sideTheme === 'theme-dark' || isDark ? '#001529' : '#fff',
        minHeight: '100vh',
        position: 'sticky',
        top: 0,
        height: '100vh',
        overflowY: 'auto',
      }}
    >
      {sidebarLogo && (
        <div style={{ height: 48, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#fff' }}>
          <span style={{ fontWeight: 600, fontSize: collapsed ? 12 : 15, whiteSpace: 'nowrap' }}>
            {collapsed ? 'RY' : '若依管理系统'}
          </span>
        </div>
      )}
      <Menu
        theme={sideTheme === 'theme-dark' || isDark ? 'dark' : 'light'}
        mode="inline"
        inlineCollapsed={collapsed}
        items={items}
        selectedKeys={[currentPath]}
        style={{ borderInlineEnd: 'none' }}
        onClick={({ key }) => {
          if (!/^https?:\/\//.test(key)) void navigate(key)
        }}
      />
    </div>
  )
}
