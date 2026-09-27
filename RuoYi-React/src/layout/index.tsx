// Layout —— 4.0.0 完整布局壳（替换 2.0.0 的临时 LayoutShell）
// 结构：Sidebar + 主区（Navbar + AppMain）；TagsView/Settings 抽屉随后续任务补

import { Outlet, useLocation } from 'react-router'
import Sidebar from './Sidebar'
import Navbar from './Navbar'
import { useAppSelector } from '@/store/hooks'

export default function Layout() {
  const sidebar = useAppSelector((s) => s.app.sidebar)
  const device = useAppSelector((s) => s.app.device)
  const location = useLocation()

  const collapse = !sidebar.opened || sidebar.hide || device === 'mobile'

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      {!sidebar.hide && <Sidebar collapsed={collapse} currentPath={location.pathname} />}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        <Navbar />
        <main style={{ flex: 1, background: '#f0f2f5' }}>
          <Outlet />
        </main>
      </div>
    </div>
  )
}
