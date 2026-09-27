// Breadcrumb —— 对位基准 components/Breadcrumb
// route.matched 过滤有 title 的项；非首页前置「首页」；redirect=noRedirect 纯文本

import { Breadcrumb as AntBreadcrumb } from 'antd'
import { Link, useLocation, matchRoutes } from 'react-router'
import { HomeOutlined } from '@ant-design/icons'
import { constantRoutes, layoutRoute } from '@/router/routes'

export default function Breadcrumb() {
  const location = useLocation()
  const allRoutes = [...constantRoutes, layoutRoute]
  const matched = matchRoutes(allRoutes as never, location.pathname) || []

  const items = [{ title: <Link to="/index"><HomeOutlined /> 首页</Link> }]
  for (const m of matched) {
    const handle = (m.route as unknown as { handle?: { title?: string; metaLink?: string } }).handle
    if (!handle?.title) continue
    if (handle.title === '首页') continue
    items.push({
      title: handle.metaLink ? (
        <span>{handle.title}</span>
      ) : (
        <Link to={m.pathnameBase || handle.title && m.route.path || '/index'}>{handle.title}</Link>
      ),
    })
  }
  return <AntBreadcrumb items={items} style={{ marginLeft: 8 }} />
}
