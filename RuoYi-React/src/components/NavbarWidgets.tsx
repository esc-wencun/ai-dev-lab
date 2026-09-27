// Hamburger / Screenfull / SizeSelect —— Navbar 右侧与开合组件（对位基准同名组件）

import { useState } from 'react'
import { MenuFoldOutlined, MenuUnfoldOutlined, FullscreenOutlined, FullscreenExitOutlined } from '@ant-design/icons'
import { Dropdown } from 'antd'
import Cookies from 'js-cookie'
import { toggleSideBar } from '@/store/modules/app'
import { useAppDispatch, useAppSelector } from '@/store/hooks'

export function Hamburger() {
  const dispatch = useAppDispatch()
  const opened = useAppSelector((s) => s.app.sidebar.opened)
  return (
    <span
      onClick={() => dispatch(toggleSideBar())}
      style={{ cursor: 'pointer', padding: '0 12px', display: 'inline-flex', alignItems: 'center' }}
    >
      {opened ? <MenuFoldOutlined /> : <MenuUnfoldOutlined />}
    </span>
  )
}

export function Screenfull() {
  const [isFull, setIsFull] = useState(false)
  const toggle = () => {
    if (!document.fullscreenElement) {
      void document.documentElement.requestFullscreen()
      setIsFull(true)
    } else {
      void document.exitFullscreen()
      setIsFull(false)
    }
  }
  return (
    <span onClick={toggle} style={{ cursor: 'pointer', padding: '0 12px', display: 'inline-flex', alignItems: 'center' }}>
      {isFull ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
    </span>
  )
}

// SizeSelect：组件尺寸切换（选择后写 Cookie + 整页刷新，对齐基准行为）
export function SizeSelect() {
  const size = useAppSelector((s) => s.app.size)
  const dispatch = useAppDispatch()
  const items = [
    { key: 'large', label: '较大' },
    { key: 'default', label: '默认' },
    { key: 'small', label: '稍小' },
  ]
  return (
    <Dropdown
      menu={{
        items,
        selectable: true,
        defaultSelectedKeys: [size],
        onClick: ({ key }) => {
          dispatch({ type: 'app/setSize', payload: key } as never)
          Cookies.set('size', key)
          window.location.reload()
        },
      }}
    >
      <span style={{ cursor: 'pointer', padding: '0 12px', display: 'inline-flex', alignItems: 'center' }}>
        布局大小
      </span>
    </Dropdown>
  )
}
