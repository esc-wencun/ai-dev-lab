// TagsView 页签栏 —— 对位基准 layout/components/TagsView/index.vue
// 路由变化 addTags / 挂载初始化 affix 页签 / 中键关闭 / 右键菜单六项 / 下拉菜单+全屏(Esc 退出) /
// 刷新走 /redirect 中转 / 左右滚动箭头+活动页签滚入视野 / card|chrome 双样式 / tagsIcon / 持久化

import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router'
import { useRouteMatches } from '@/hooks/useRouteMatches'
import type { CSSProperties, MouseEvent } from 'react'
import { Dropdown } from 'antd'
import type { MenuProps } from 'antd'
import {
  ArrowLeftOutlined,
  ArrowRightOutlined,
  ArrowDownOutlined,
  CloseOutlined,
  ReloadOutlined,
  VerticalLeftOutlined,
  VerticalRightOutlined,
  FullscreenOutlined,
} from '@ant-design/icons'
import SvgIcon from '@/components/SvgIcon'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import {
  addAffixView,
  addView,
  loadPersistedViews,
  updateVisitedView,
  type TagView,
} from '@/store/modules/tagsView'
import { useTabActions } from '@/hooks/useTabActions'
import { getNormalPath } from '@/utils/ruoyi'
import { store } from '@/store'
import { layoutRoute } from '@/router/routes'

// affix 提取的最小路由形状（layoutRoute.children 的 handle 形态 / RouteItem 的 meta 形态均满足）
interface AffixRouteLike {
  name?: string
  path: string
  hidden?: boolean
  meta?: { affix?: boolean; title?: string; [k: string]: unknown }
  handle?: { name?: string; title?: string; affix?: boolean; [k: string]: unknown }
  children?: AffixRouteLike[]
}

// 从路由表提取 affix 页签（对位基准 filterAffixTags 递归）
// 静态路由 affix 在 handle 上（routes.tsx），后端菜单 affix 在 meta 上
function filterAffixTags(routes: AffixRouteLike[], basePath = ''): TagView[] {
  let tags: TagView[] = []
  routes.forEach((route) => {
    if (route.hidden) return
    if (route.meta && route.meta.affix) {
      const tagPath = getNormalPath(basePath + '/' + route.path)
      tags.push({
        fullPath: tagPath,
        path: tagPath,
        name: route.name,
        title: route.meta.title,
        meta: { ...route.meta },
      })
    } else if (route.handle && route.handle.affix) {
      const tagPath = getNormalPath(basePath + '/' + route.path)
      tags.push({
        fullPath: tagPath,
        path: tagPath,
        name: route.handle.name || route.name,
        title: route.handle.title,
        meta: { affix: true },
      })
    }
    if (route.children) {
      const tempTags = filterAffixTags(route.children, route.path)
      if (tempTags.length >= 1) tags = [...tags, ...tempTags]
    }
  })
  return tags
}

export default function TagsView() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const location = useLocation()
  const matches = useRouteMatches()
  const { refreshPage, closePage, closeOtherPage, closeLeftPage, closeRightPage, closeAllPage } = useTabActions()

  const visitedViews = useAppSelector((s) => s.tagsView.visitedViews)
  const sidebarRoutes = useAppSelector((s) => s.permission.sidebarRoutes)
  const theme = useAppSelector((s) => s.settings.theme)
  const tagsIcon = useAppSelector((s) => s.settings.tagsIcon)
  const tagsViewPersist = useAppSelector((s) => s.settings.tagsViewPersist)
  const tagsViewStyle = useAppSelector((s) => s.settings.tagsViewStyle)

  const [contextMenu, setContextMenu] = useState<{ visible: boolean; left: number; top: number; tag?: TagView }>({ visible: false, left: 0, top: 0 })
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [canLeft, setCanLeft] = useState(false)
  const [canRight, setCanRight] = useState(false)
  const affixTagsRef = useRef<TagView[]>([])
  const scrollRef = useRef<HTMLDivElement>(null)
  const initedRef = useRef(false)

  const currentFullPath = location.pathname + location.search
  // 当前激活 tag；下拉菜单针对当前激活的 tag（对位 selectedDropdownTag）
  const activeTag = visitedViews.find((v) => v.path === location.pathname)
  const selectedDropdownTag = activeTag || visitedViews[0]

  const isActive = useCallback((r: TagView) => r.path === location.pathname, [location.pathname])

  const isAffix = useCallback((tag?: TagView) => !!(tag && tag.meta && tag.meta.affix), [])

  // ---------- 滚动控制（简化版 ScrollPane：overflow 容器 + scrollTo） ----------

  const updateArrowState = useCallback(() => {
    const el = scrollRef.current
    if (!el) return
    setCanLeft(el.scrollLeft > 0)
    setCanRight(el.scrollLeft < el.scrollWidth - el.clientWidth - 1)
  }, [])

  // 活动页签滚入视野（简化 moveToTarget）
  const moveToCurrentTag = useCallback(() => {
    const el = scrollRef.current
    if (!el) return
    const active = el.querySelector<HTMLElement>('.tags-view-item.active')
    if (active) {
      const offsetLeft = active.offsetLeft - el.offsetLeft
      const offsetWidth = active.offsetWidth
      if (offsetLeft < el.scrollLeft) el.scrollTo({ left: offsetLeft - 10, behavior: 'smooth' })
      else if (offsetLeft + offsetWidth > el.scrollLeft + el.clientWidth) {
        el.scrollTo({ left: offsetLeft + offsetWidth - el.clientWidth + 10, behavior: 'smooth' })
      }
    }
  }, [])

  // ---------- 页签维护 ----------

  // 路由变化 → addTags + fullPath 维护（对位基准 watch(route) → addTags + moveToCurrentTag）
  useEffect(() => {
    const leaf = [...matches].reverse().find((m) => m.handle)
    // handle 归一：静态路由（routes.tsx）affix 平铺在 handle 上，动态路由在 handle.meta
    const rawHandle = (leaf?.handle || {}) as { name?: string; title?: string; meta?: TagView['meta']; affix?: boolean }
    const handle = {
      name: rawHandle.name,
      title: rawHandle.title,
      meta: { ...(rawHandle.meta || {}), ...(rawHandle.affix ? { affix: true } : {}) } as TagView['meta'],
    }
    const query = Object.fromEntries(new URLSearchParams(location.search))
    if (handle.name) {
      dispatch(
        addView({
          name: handle.name,
          path: location.pathname,
          fullPath: currentFullPath,
          title: handle.title,
          query,
          meta: handle.meta,
        }),
      )
      // query 变化时更新既有页签（对位 updateVisitedView）
      const existing = store.getState().tagsView.visitedViews.find((v) => v.path === location.pathname)
      if (existing && existing.fullPath !== currentFullPath) {
        dispatch(updateVisitedView({ name: handle.name, path: location.pathname, fullPath: currentFullPath, title: handle.title, query, meta: handle.meta }))
      }
    }
    moveToCurrentTag()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.pathname, location.search])

  // 挂载初始化：persist 恢复 + affix 页签 unshift（对位 initTags；/index 是 affix）
  useEffect(() => {
    if (initedRef.current) return
    initedRef.current = true
    if (tagsViewPersist) {
      try {
        const saved = localStorage.getItem('tags-view-visited')
        if (saved) dispatch(loadPersistedViews(JSON.parse(saved)))
      } catch {
        /* ignore */
      }
    }
    // 静态布局子路由（/index affix）+ 后端菜单（sidebarRoutes 可能含 affix 项）
    const res = [
      ...filterAffixTags(layoutRoute.children as unknown as AffixRouteLike[]),
      ...filterAffixTags(sidebarRoutes as unknown as AffixRouteLike[]),
    ]
    affixTagsRef.current = res
    for (const tag of res) {
      if (tag.name) dispatch(addAffixView(tag))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // visitedViews 变化 → 箭头状态 + 持久化（对位 saveVisitedViews：排除 affix，存六字段）
  useEffect(() => {
    updateArrowState()
    if (tagsViewPersist) {
      const toSave = visitedViews
        .filter((v) => !(v.meta && v.meta.affix))
        .map((v) => ({ path: v.path, fullPath: v.fullPath, name: v.name, title: v.title, query: v.query, meta: v.meta }))
      try {
        localStorage.setItem('tags-view-visited', JSON.stringify(toSave))
      } catch {
        /* ignore */
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visitedViews, tagsViewPersist])

  // persist 关闭时删除持久化键（对位基准：persist 关 → 删 tags-view-visited）
  useEffect(() => {
    if (!tagsViewPersist) {
      try {
        localStorage.removeItem('tags-view-visited')
      } catch {
        /* ignore */
      }
    }
  }, [tagsViewPersist])

  // ---------- 全屏（对位 toggleFullscreen：隐藏 Navbar+Sidebar，Esc 退出） ----------

  const toggleFullscreen = useCallback(() => {
    const navbar = document.querySelector('.navbar')
    const sidebar = document.querySelector('.sidebar-container')
    const mainContainer = document.querySelector('.main-container')
    if (!mainContainer) return
    if (!isFullscreen) {
      mainContainer.classList.add('fullscreen-mode')
      document.body.style.overflow = 'hidden'
      if (navbar) (navbar as HTMLElement).style.display = 'none'
      if (sidebar) (sidebar as HTMLElement).style.display = 'none'
      setIsFullscreen(true)
    } else {
      mainContainer.classList.remove('fullscreen-mode')
      document.body.style.overflow = ''
      if (navbar) (navbar as HTMLElement).style.display = ''
      if (sidebar) (sidebar as HTMLElement).style.display = ''
      setIsFullscreen(false)
    }
  }, [isFullscreen])

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isFullscreen) toggleFullscreen()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [isFullscreen, toggleFullscreen])

  useEffect(() => {
    window.addEventListener('resize', updateArrowState)
    return () => window.removeEventListener('resize', updateArrowState)
  }, [updateArrowState])

  // ---------- 右键菜单 ----------

  const closeMenu = useCallback(() => {
    setContextMenu((s) => ({ ...s, visible: false }))
    document.body.removeEventListener('click', closeMenu)
  }, [])

  const openMenu = (tag: TagView, e: MouseEvent) => {
    e.preventDefault()
    setContextMenu({ visible: true, left: e.clientX, top: e.clientY, tag })
    document.body.addEventListener('click', closeMenu)
  }

  // ---------- 关闭动作（跳转语义对位基准） ----------

  const toLastView = (views: TagView[], view?: TagView) => {
    const latest = views[views.length - 1]
    if (latest) {
      void navigate(latest.fullPath || latest.path)
    } else if (view && view.name === 'Dashboard') {
      navigate({ pathname: '/redirect' + view.fullPath }, { replace: true })
    } else {
      void navigate('/')
    }
  }

  const handleCloseSelected = (view: TagView) => {
    closePage(view)
    // dispatch 同步，state 已更新
    const views = store.getState().tagsView.visitedViews
    if (isActive(view)) toLastView(views, view)
  }

  const handleCloseRight = (view: TagView) => {
    closeRightPage(view)
    const views = store.getState().tagsView.visitedViews
    if (!views.find((i) => i.fullPath === currentFullPath)) toLastView(views)
  }

  const handleCloseLeft = (view: TagView) => {
    closeLeftPage(view)
    const views = store.getState().tagsView.visitedViews
    if (!views.find((i) => i.fullPath === currentFullPath)) toLastView(views)
  }

  const handleCloseOthers = (view: TagView) => {
    void navigate(view.fullPath || view.path)
    closeOtherPage(view)
    moveToCurrentTag()
  }

  const handleCloseAll = (view?: TagView) => {
    closeAllPage()
    const views = store.getState().tagsView.visitedViews
    if (affixTagsRef.current.some((tag) => tag.path === location.pathname)) return
    toLastView(views, view)
  }

  const handleRefresh = (view?: TagView) => {
    refreshPage(view)
  }

  function isFirstView(tag?: TagView): boolean {
    try {
      return tag?.fullPath === '/index' || tag?.fullPath === visitedViews[1].fullPath
    } catch {
      return false
    }
  }

  function isLastView(tag?: TagView): boolean {
    try {
      return tag?.fullPath === visitedViews[visitedViews.length - 1].fullPath
    } catch {
      return false
    }
  }

  // 下拉菜单（当前激活页签）：关闭当前/其他/左侧/右侧/全部 + 全屏切换
  const dropdownItems: MenuProps['items'] = [
    !isAffix(selectedDropdownTag) && { key: 'close', icon: <CloseOutlined />, label: '关闭当前' },
    { key: 'closeOthers', icon: <CloseOutlined />, label: '关闭其他' },
    { key: 'closeLeft', icon: <VerticalLeftOutlined />, label: '关闭左侧', disabled: isFirstView(selectedDropdownTag) },
    { key: 'closeRight', icon: <VerticalRightOutlined />, label: '关闭右侧', disabled: isLastView(selectedDropdownTag) },
    { key: 'closeAll', icon: <CloseOutlined />, label: '全部关闭' },
    { type: 'divider' },
    { key: 'fullscreen', icon: <FullscreenOutlined />, label: isFullscreen ? '退出全屏' : '全屏显示' },
  ].filter(Boolean) as MenuProps['items']

  const handleDropdownClick: MenuProps['onClick'] = ({ key }) => {
    const tag = selectedDropdownTag
    if (!tag) return
    switch (key) {
      case 'close': handleCloseSelected(tag); break
      case 'closeOthers': handleCloseOthers(tag); break
      case 'closeLeft': handleCloseLeft(tag); break
      case 'closeRight': handleCloseRight(tag); break
      case 'closeAll': handleCloseAll(tag); break
      case 'fullscreen': toggleFullscreen(); break
    }
  }

  // card 模式激活页签背景 = settings.theme（对位 tagActiveStyle）
  const tagActiveStyle = (tag: TagView): CSSProperties => {
    if (!isActive(tag) || tagsViewStyle !== 'card') return {}
    return { backgroundColor: theme, borderColor: theme }
  }

  const scrollByStep = (dir: number) => {
    scrollRef.current?.scrollBy({ left: dir * 200, behavior: 'smooth' })
  }

  return (
    <div id="tags-view-container" className={`tags-view-container${tagsViewStyle === 'chrome' ? ' tags-view-container--chrome' : ''}`}>
      {/* 左切换箭头 */}
      <span className={`tags-nav-btn tags-nav-btn--left${canLeft ? '' : ' disabled'}`} onClick={() => scrollByStep(-1)}>
        <ArrowLeftOutlined />
      </span>

      {/* 标签滚动区 */}
      <div
        className="tags-view-wrapper"
        ref={scrollRef}
        onScroll={() => { closeMenu(); updateArrowState() }}
      >
        {visitedViews.map((tag) => (
          <Link
            key={tag.path}
            data-path={tag.path}
            to={tag.fullPath || tag.path}
            className={`tags-view-item${isActive(tag) ? ' active' : ''}${tagsIcon ? ' has-icon' : ''}`}
            style={tagActiveStyle(tag)}
            onMouseDown={(e) => {
              // 中键点击关闭（非 affix）
              if (e.button === 1 && !isAffix(tag)) {
                e.preventDefault()
                handleCloseSelected(tag)
              }
            }}
            onContextMenu={(e) => openMenu(tag, e)}
          >
            {tagsIcon && typeof tag.meta?.icon === 'string' && tag.meta.icon !== '#' && (
              <SvgIcon iconClass={tag.meta.icon} size={12} className="tags-item-icon" />
            )}
            {tag.title || 'no-name'}
            {!isAffix(tag) && (
              <span
                className="tags-close-btn"
                onClick={(e) => { e.preventDefault(); e.stopPropagation(); handleCloseSelected(tag) }}
              >
                <CloseOutlined style={{ fontSize: 10 }} />
              </span>
            )}
          </Link>
        ))}
      </div>

      {/* 右切换箭头 */}
      <span className={`tags-nav-btn tags-nav-btn--right${canRight ? '' : ' disabled'}`} onClick={() => scrollByStep(1)}>
        <ArrowRightOutlined />
      </span>

      {/* 下拉操作菜单 */}
      <Dropdown menu={{ items: dropdownItems, onClick: handleDropdownClick }} trigger={['click']} placement="bottomRight">
        <span className="tags-action-btn">
          <ArrowDownOutlined />
        </span>
      </Dropdown>

      {/* 刷新按钮 */}
      <span className="tags-action-btn tags-refresh-btn" title="刷新页面" onClick={() => handleRefresh(selectedDropdownTag)}>
        <ReloadOutlined /> 刷新
      </span>

      {/* 右键上下文菜单 */}
      {contextMenu.visible && contextMenu.tag && (() => {
        const tag = contextMenu.tag
        return (
          <ul className="contextmenu" style={{ left: contextMenu.left, top: contextMenu.top }}>
            <li onClick={() => { handleRefresh(tag); closeMenu() }}><ReloadOutlined /> 刷新页面</li>
            {!isAffix(tag) && (
              <li onClick={() => { handleCloseSelected(tag); closeMenu() }}><CloseOutlined /> 关闭当前</li>
            )}
            <li onClick={() => { handleCloseOthers(tag); closeMenu() }}><CloseOutlined /> 关闭其他</li>
            {!isFirstView(tag) && (
              <li onClick={() => { handleCloseLeft(tag); closeMenu() }}><VerticalLeftOutlined /> 关闭左侧</li>
            )}
            {!isLastView(tag) && (
              <li onClick={() => { handleCloseRight(tag); closeMenu() }}><VerticalRightOutlined /> 关闭右侧</li>
            )}
            <li onClick={() => { handleCloseAll(tag); closeMenu() }}><CloseOutlined /> 全部关闭</li>
          </ul>
        )
      })()}
    </div>
  )
}
