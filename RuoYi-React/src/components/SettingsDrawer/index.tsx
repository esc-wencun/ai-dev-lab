// SettingsDrawer 布局设置抽屉 —— 对位基准 layout/components/Settings/index.vue
// 11 项：导航模式(1/2/3 示意图) / 侧边栏主题(dark/light) / 主题颜色(ColorPicker 色板) /
// 开启页签 / 持久化标签页(页签关时禁用) / 显示页签图标 / 标签页样式(卡片/谷歌) /
// 固定 Header / 显示 Logo / 动态标题 / 底部版权
// 保存：loading 提示 + 11 字段 JSON 存 localStorage `layout-setting`；persist 关时删 `tags-view-visited`
// 重置：删两键 + location.reload()

import { App, Button, ColorPicker, Divider, Drawer, Radio, Space, Switch, Tooltip } from 'antd'
import type { Color } from 'antd/es/color-picker'
import { CheckOutlined, FileAddOutlined, RedoOutlined } from '@ant-design/icons'
import { changeSetting, setTitle } from '@/store/modules/settings'
import { closeSideBar, toggleSideBar, toggleSideBarHide } from '@/store/modules/app'
import { setSidebarRoutes } from '@/store/modules/permission'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import darkSvg from '@/assets/images/dark.svg'
import lightSvg from '@/assets/images/light.svg'

// 预置色板（与基准 predefineColors 一致）
const PREDEFINE_COLORS = ['#409EFF', '#ff4500', '#ff8c00', '#ffd700', '#90ee90', '#00ced1', '#1e90ff', '#c71585']

// localStorage 键名（契约）
const LAYOUT_KEY = 'layout-setting'
const TAGS_KEY = 'tags-view-visited'

export default function SettingsDrawer({ open, onClose }: { open: boolean; onClose: () => void }) {
  const dispatch = useAppDispatch()
  const { message } = App.useApp()
  const settings = useAppSelector((s) => s.settings)
  const sidebarOpened = useAppSelector((s) => s.app.sidebar.opened)
  const defaultRoutes = useAppSelector((s) => s.permission.defaultRoutes)

  // 开关项通用 changeSetting
  const setBool = (key: 'tagsView' | 'tagsViewPersist' | 'tagsIcon' | 'fixedHeader' | 'sidebarLogo' | 'dynamicTitle' | 'footerVisible', value: boolean) => {
    dispatch(changeSetting({ key, value }))
  }

  // 导航模式切换副作用（对位基准 watch navType）：
  // 1 → opened=true + hide=false；2 → opened=true；3 → opened=false + hide=true；1/3 复位侧边栏为默认菜单
  const handleNavType = (val: 1 | 2 | 3) => {
    dispatch(changeSetting({ key: 'navType', value: val }))
    if (val === 1) {
      if (!sidebarOpened) dispatch(toggleSideBar())
      dispatch(toggleSideBarHide(false))
    }
    if (val === 2) {
      if (!sidebarOpened) dispatch(toggleSideBar())
    }
    if (val === 3) {
      dispatch(closeSideBar({ withoutAnimation: false }))
      dispatch(toggleSideBarHide(true))
    }
    if (val === 1 || val === 3) {
      dispatch(setSidebarRoutes(defaultRoutes))
    }
  }

  // 侧边栏主题
  const handleSideTheme = (val: 'theme-dark' | 'theme-light') => {
    dispatch(changeSetting({ key: 'sideTheme', value: val }))
  }

  // 主题颜色（antd ColorPicker → hex 字符串）
  const themeChange = (color: Color | string) => {
    const hex = typeof color === 'string' ? color : color.toHexString()
    dispatch(changeSetting({ key: 'theme', value: hex }))
  }

  // 动态标题切换（对位 dynamicTitleChange：setTitle 重算 document.title）
  const dynamicTitleChange = (v: boolean) => {
    setBool('dynamicTitle', v)
    dispatch(setTitle(settings.title))
  }

  // 保存配置（对位 saveSetting）
  const saveSetting = () => {
    message.loading({ content: '正在保存到本地，请稍候...', key: 'save-setting', duration: 0 })
    if (!settings.tagsViewPersist) {
      localStorage.removeItem(TAGS_KEY)
    }
    const layoutSetting = {
      navType: settings.navType,
      tagsView: settings.tagsView,
      tagsIcon: settings.tagsIcon,
      tagsViewStyle: settings.tagsViewStyle,
      tagsViewPersist: settings.tagsViewPersist,
      fixedHeader: settings.fixedHeader,
      sidebarLogo: settings.sidebarLogo,
      dynamicTitle: settings.dynamicTitle,
      footerVisible: settings.footerVisible,
      sideTheme: settings.sideTheme,
      theme: settings.theme,
    }
    try {
      localStorage.setItem(LAYOUT_KEY, JSON.stringify(layoutSetting))
    } catch {
      /* ignore */
    }
    setTimeout(() => message.destroy('save-setting'), 1000)
  }

  // 重置配置（对位 resetSetting：删两键 + 整页刷新）
  const resetSetting = () => {
    localStorage.removeItem(TAGS_KEY)
    message.loading({ content: '正在清除设置缓存并刷新，请稍候...', key: 'reset-setting', duration: 0 })
    localStorage.removeItem(LAYOUT_KEY)
    setTimeout(() => window.location.reload(), 1000)
  }

  // 通用开关行
  const renderSwitch = (
    label: string,
    key: 'tagsView' | 'tagsViewPersist' | 'tagsIcon' | 'fixedHeader' | 'sidebarLogo' | 'dynamicTitle' | 'footerVisible',
    opts?: { disabled?: boolean; onChange?: (v: boolean) => void },
  ) => (
    <div className="drawer-item">
      <span>{label}</span>
      <span className="comp-style">
        <Switch
          checked={settings[key] as boolean}
          disabled={opts?.disabled}
          onChange={(v) => {
            setBool(key, v)
            opts?.onChange?.(v)
          }}
        />
      </span>
    </div>
  )

  return (
    <Drawer open={open} onClose={onClose} title={null} width={300} styles={{ body: { padding: '16px 20px' } }}>
      <div className="setting-drawer-title">
        <h3 className="drawer-title">菜单导航设置</h3>
      </div>
      {/* 导航模式：1/2/3 三张 CSS 示意图 */}
      <div className="nav-wrap">
        <Tooltip title="左侧菜单" placement="bottom">
          <div className={`item left${settings.navType === 1 ? ' activeItem' : ''}`} onClick={() => handleNavType(1)}>
            <b />
            <b />
          </div>
        </Tooltip>
        <Tooltip title="混合菜单" placement="bottom">
          <div className={`item mix${settings.navType === 2 ? ' activeItem' : ''}`} onClick={() => handleNavType(2)}>
            <b />
            <b />
          </div>
        </Tooltip>
        <Tooltip title="顶部菜单" placement="bottom">
          <div className={`item top${settings.navType === 3 ? ' activeItem' : ''}`} onClick={() => handleNavType(3)}>
            <b />
            <b />
          </div>
        </Tooltip>
      </div>

      <div className="setting-drawer-title">
        <h3 className="drawer-title">主题风格设置</h3>
      </div>
      {/* 侧边栏主题：dark / light 两图 */}
      <div className="setting-drawer-block-checbox">
        <div className="setting-drawer-block-checbox-item" onClick={() => handleSideTheme('theme-dark')}>
          <img src={darkSvg} alt="dark" />
          {settings.sideTheme === 'theme-dark' && (
            <div className="setting-drawer-block-checbox-selectIcon">
              <CheckOutlined style={{ color: settings.theme }} />
            </div>
          )}
        </div>
        <div className="setting-drawer-block-checbox-item" onClick={() => handleSideTheme('theme-light')}>
          <img src={lightSvg} alt="light" />
          {settings.sideTheme === 'theme-light' && (
            <div className="setting-drawer-block-checbox-selectIcon">
              <CheckOutlined style={{ color: settings.theme }} />
            </div>
          )}
        </div>
      </div>
      {/* 主题颜色 */}
      <div className="drawer-item">
        <span>主题颜色</span>
        <span className="comp-style">
          <ColorPicker
            value={settings.theme}
            presets={[{ label: '预置色板', colors: PREDEFINE_COLORS }]}
            onChange={themeChange}
            disabledAlpha
          />
        </span>
      </div>

      <Divider />

      <h3 className="drawer-title">系统布局配置</h3>

      {renderSwitch('开启页签', 'tagsView')}
      {renderSwitch('持久化标签页', 'tagsViewPersist', { disabled: !settings.tagsView })}
      {renderSwitch('显示页签图标', 'tagsIcon', { disabled: !settings.tagsView })}
      {/* 标签页样式：卡片 / 谷歌 radio（页签关时禁用） */}
      <div className="drawer-item">
        <span>标签页样式</span>
        <span className="comp-style">
          <Radio.Group
            value={settings.tagsViewStyle}
            disabled={!settings.tagsView}
            size="small"
            onChange={(e) => dispatch(changeSetting({ key: 'tagsViewStyle', value: e.target.value }))}
          >
            <Radio.Button value="card">卡片</Radio.Button>
            <Radio.Button value="chrome">谷歌</Radio.Button>
          </Radio.Group>
        </span>
      </div>
      {renderSwitch('固定 Header', 'fixedHeader')}
      {renderSwitch('显示 Logo', 'sidebarLogo')}
      {renderSwitch('动态标题', 'dynamicTitle', { onChange: dynamicTitleChange })}
      {renderSwitch('底部版权', 'footerVisible')}

      <Divider />

      <Space>
        <Button type="primary" ghost icon={<FileAddOutlined />} onClick={saveSetting}>
          保存配置
        </Button>
        <Button icon={<RedoOutlined />} onClick={resetSetting}>
          重置配置
        </Button>
      </Space>
    </Drawer>
  )
}
