// Navbar —— 对位基准 layout/components/Navbar（navType 1 形态为主，2/3 随任务补）
// 右侧组件条：HeaderSearch / 全屏 / 明暗切换 / 尺寸 / 锁屏 / 头像下拉（个人中心/退出）

import { useNavigate } from 'react-router'
import { Avatar, Dropdown } from 'antd'
import { UserOutlined, BulbOutlined, LockOutlined, LogoutOutlined } from '@ant-design/icons'
import Breadcrumb from '@/components/Breadcrumb'
import HeaderSearch from '@/components/HeaderSearch'
import { Hamburger, Screenfull, SizeSelect } from '@/components/NavbarWidgets'
import { logOut } from '@/store/modules/user'
import { lockScreen } from '@/store/modules/lock'
import { toggleTheme } from '@/store/modules/settings'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import { Modal } from 'antd'

export default function Navbar() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const nickName = useAppSelector((s) => s.user.nickName)
  const avatar = useAppSelector((s) => s.user.avatar)

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
        navigate('/login', { replace: true })
      },
    })
  }

  return (
    <div
      style={{
        height: 48, display: 'flex', alignItems: 'center', background: '#fff',
        borderBottom: '1px solid #e8e8e8', position: 'sticky', top: 0, zIndex: 10,
      }}
    >
      <Hamburger />
      <Breadcrumb />
      <div style={{ flex: 1 }} />
      <div style={{ display: 'flex', alignItems: 'center', height: '100%' }}>
        <HeaderSearch />
        <Screenfull />
        <SizeSelect />
        <span onClick={() => dispatch(toggleTheme())} style={{ cursor: 'pointer', padding: '0 12px', display: 'inline-flex' }}>
          <BulbOutlined />
        </span>
        <span onClick={handleLock} style={{ cursor: 'pointer', padding: '0 12px', display: 'inline-flex' }}>
          <LockOutlined />
        </span>
        <Dropdown
          menu={{
            items: [
              { key: 'profile', icon: <UserOutlined />, label: '个人中心' },
              { key: 'lock', icon: <LockOutlined />, label: '锁定屏幕' },
              { type: 'divider' },
              { key: 'logout', icon: <LogoutOutlined />, label: '退出登录' },
            ],
            onClick: ({ key }) => {
              if (key === 'profile') void navigate('/user/profile')
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
