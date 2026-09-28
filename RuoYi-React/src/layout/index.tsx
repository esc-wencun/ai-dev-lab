// Layout —— 4.0.0 完整布局壳（对位基准 layout/index.vue）
// Sidebar + 主区（Navbar + TagsView + AppMain + Copyright footer）
// fixedHeader 语义 / 响应式 <992 切 mobile + 遮罩 / InnerLink iframe 页渲染 / 全屏模式

import { useEffect, useState } from 'react'
import { Outlet, useLocation } from 'react-router'
import Sidebar from './Sidebar'
import Navbar from './Navbar'
import TagsView from '@/components/TagsView'
import SettingsDrawer from '@/components/SettingsDrawer'
import IFrame from '@/components/IFrame'
import { useRouteMatches } from '@/hooks/useRouteMatches'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import { closeSideBar, toggleDevice } from '@/store/modules/app'

const WIDTH = 992 // refer to Bootstrap's responsive design

export default function Layout() {
  const dispatch = useAppDispatch()
  const location = useLocation()
  const matches = useRouteMatches()

  const sidebar = useAppSelector((s) => s.app.sidebar)
  const device = useAppSelector((s) => s.app.device)
  const theme = useAppSelector((s) => s.settings.theme)
  const tagsView = useAppSelector((s) => s.settings.tagsView)
  const fixedHeader = useAppSelector((s) => s.settings.fixedHeader)
  const footerVisible = useAppSelector((s) => s.settings.footerVisible)
  const footerContent = useAppSelector((s) => s.settings.footerContent)

  // 设置抽屉开关（Navbar 头像下拉「布局设置」打开）
  const [settingsOpen, setSettingsOpen] = useState(false)

  // 响应式（对位基准 watchEffect：宽 <992 → mobile + 收起侧栏）
  useEffect(() => {
    const onResize = () => {
      if (window.innerWidth - 1 < WIDTH) {
        dispatch(toggleDevice('mobile'))
        dispatch(closeSideBar({ withoutAnimation: true }))
      } else {
        dispatch(toggleDevice('desktop'))
      }
    }
    onResize()
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [dispatch])

  // 当前路由是否 InnerLink（metaLink 存在 → iframe 渲染替代 Outlet）
  const leafHandle = [...matches].reverse().find((m) => m.handle)?.handle as
    | { metaLink?: string | null }
    | undefined
  const metaLink = leafHandle?.metaLink || null

  // iframe src = metaLink + query 拼接（对位 IframeToggle.iframeUrl）
  const iframeSrc = metaLink
    ? metaLink + (location.search || '')
    : null

  const collapse = !sidebar.opened || sidebar.hide || device === 'mobile'
  const showSidebar = !sidebar.hide

  return (
    <div className={`app-wrapper${!sidebar.opened ? ' hideSidebar' : ' openSidebar'}${device === 'mobile' ? ' mobile' : ''}${sidebar.withoutAnimation ? ' withoutAnimation' : ''}`}>
      {/* 移动端遮罩（opened 时渲染，点击关闭侧栏） */}
      {device === 'mobile' && sidebar.opened && (
        <div className="drawer-bg" onClick={() => dispatch(closeSideBar({ withoutAnimation: false }))} />
      )}

      {showSidebar && <Sidebar collapsed={collapse} currentPath={location.pathname} />}

      <div className={`main-container${tagsView ? ' hasTagsView' : ''}${sidebar.hide ? ' sidebarHide' : ''}`}>
        <div className={fixedHeader ? 'fixed-header' : ''}>
          <Navbar onOpenSettings={() => setSettingsOpen(true)} />
          {tagsView && <TagsView />}
        </div>
        <main className="app-main">
          {iframeSrc ? (
            // InnerLink 页：全高 iframe 替代 Outlet（对位 IframeToggle/InnerLink）
            <IFrame key={iframeSrc} src={iframeSrc} />
          ) : (
            <Outlet />
          )}
        </main>
        {/* footer（对位 Copyright：36px 底条，footerVisible 控制） */}
        {footerVisible && (
          <footer className="copyright">
            <span>{footerContent}</span>
          </footer>
        )}
      </div>

      <SettingsDrawer open={settingsOpen} onClose={() => setSettingsOpen(false)} />
      {/* 主题色 CSS 变量（对位 app-wrapper 的 --current-color 注入） */}
      <style>{`:root{--ry-current-color:${theme}}`}</style>
    </div>
  )
}
