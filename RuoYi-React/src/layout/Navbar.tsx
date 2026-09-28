// Navbar —— 对位基准 layout/components/Navbar.vue
// navType 1：hamburger + breadcrumb；2：hamburger + TopNav；3：Logo + TopBar（隐藏 hamburger）
// 右侧组件条：HeaderSearch / Git / Doc / 全屏 / 明暗切换 / 尺寸 / HeaderNotice / 头像下拉
// 头像下拉：个人中心 / 布局设置(showSettings) / 锁定屏幕 / 退出登录

import { useNavigate } from 'react-router'
import { Avatar, Dropdown, Tooltip } from 'antd'
import {
  UserOutlined,
  LockOutlined,
  LogoutOutlined,
  SettingOutlined,
  MoonOutlined,
  SunOutlined,
  GithubOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons'
import Breadcrumb from '@/components/Breadcrumb'
import HeaderSearch from '@/components/HeaderSearch'
import TopNav from '@/components/TopNav'
import TopBar from '@/components/TopBar'
import HeaderNotice from '@/components/HeaderNotice'
import { Hamburger, Screenfull, SizeSelect } from '@/components/NavbarWidgets'
import { logOut } from '@/store/modules/user'
import { lockScreen } from '@/store/modules/lock'
import { toggleTheme } from '@/store/modules/settings'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import { Modal } from 'antd'

export default function Navbar({ onOpenSettings }: { onOpenSettings: () => void }) {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const nickName = useAppSelector((s) => s.user.nickName)
  const avatar = useAppSelector((s) => s.user.avatar)
  const navType = useAppSelector((s) => s.settings.navType)
  const isDark = useAppSelector((s) => s.settings.isDark)
  const showSettings = useAppSelector((s) => s.settings.showSettings)
  const sidebarLogo = useAppSelector((s) => s.settings.sidebarLogo)
  const device = useAppSelector((s) => s.app.device)

  const handleLock = () => {
    dispatch(lockScreen(window.location.pathname + window.location.search))
    void navigate('/lock')
  }

  const handleLogout = () => {
    Modal.confirm({
      title: '系统提示',
      content: '确定注销并退出系统吗？',
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        await dispatch(logOut())
        // 对位基准 location.href = '/index'
        window.location.href = '/index'
      },
    })
  }

  return (
    <div className={`navbar nav${navType}`}>
      {/* navType 1：hamburger + 面包屑；2：hamburger + TopNav；3：Logo + TopBar（hamburger 隐藏，对位 .navbar.nav3） */}
      {navType !== 3 && <Hamburger />}
      {navType === 1 && (
        <div className="breadcrumb-container">
          <Breadcrumb />
        </div>
      )}
      {navType === 2 && (
        <div className="topmenu-container">
          <TopNav />
        </div>
      )}
      {navType === 3 && (
        <>
          {sidebarLogo && <div className="navbar-logo">若依管理系统</div>}
          <div className="topbar-container">
            <TopBar />
          </div>
        </>
      )}

      <div className="right-menu">
        {device !== 'mobile' && (
          <>
            <HeaderSearch />
            <Tooltip title="源码地址" placement="bottom">
              <span className="right-menu-item hover-effect" onClick={() => window.open('https://gitee.com/y_project/RuoYi-Vue3', '_blank')}>
                <GithubOutlined />
              </span>
            </Tooltip>
            <Tooltip title="文档地址" placement="bottom">
              <span className="right-menu-item hover-effect" onClick={() => window.open('http://doc.ruoyi.vip', '_blank')}>
                <QuestionCircleOutlined />
              </span>
            </Tooltip>
            <Screenfull />
            {/* 主题切换：isDark 显示太阳（切亮），否则月亮（切暗）——对位基准 sunny/moon */}
            <Tooltip title="主题模式" placement="bottom">
              <span className="right-menu-item hover-effect" onClick={() => dispatch(toggleTheme())}>
                {isDark ? <SunOutlined /> : <MoonOutlined />}
              </span>
            </Tooltip>
            <SizeSelect />
            <Tooltip title="消息通知" placement="bottom">
              <span className="right-menu-item hover-effect">
                <HeaderNotice />
              </span>
            </Tooltip>
          </>
        )}

        <Dropdown
          menu={{
            items: [
              { key: 'profile', icon: <UserOutlined />, label: '个人中心' },
              ...(showSettings ? [{ key: 'setLayout', icon: <SettingOutlined />, label: '布局设置' }] : []),
              { key: 'lock', icon: <LockOutlined />, label: '锁定屏幕' },
              { type: 'divider' },
              { key: 'logout', icon: <LogoutOutlined />, label: '退出登录' },
            ],
            onClick: ({ key }) => {
              if (key === 'profile') void navigate('/user/profile')
              if (key === 'setLayout') onOpenSettings()
              if (key === 'lock') handleLock()
              if (key === 'logout') handleLogout()
            },
          }}
        >
          <span style={{ cursor: 'pointer', padding: '0 16px', display: 'inline-flex', alignItems: 'center', gap: 6 }}>
            <Avatar size={26} src={avatar || undefined} icon={!avatar && <UserOutlined />} />
            {nickName}
          </span>
        </Dropdown>
      </div>
    </div>
  )
}
