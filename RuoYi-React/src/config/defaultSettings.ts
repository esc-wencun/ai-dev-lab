// 默认布局配置 —— 对位 RuoYi-Vue3 src/settings.js（默认值一致）
// 可被 localStorage 键 layout-setting 的 JSON 覆盖（settings slice 实现）

export interface AppSettings {
  title: string
  theme: string
  sideTheme: 'theme-dark' | 'theme-light'
  showSettings: boolean
  navType: 1 | 2 | 3
  tagsView: boolean
  tagsViewPersist: boolean
  tagsIcon: boolean
  tagsViewStyle: 'card' | 'chrome'
  fixedHeader: boolean
  sidebarLogo: boolean
  dynamicTitle: boolean
  footerVisible: boolean
  footerContent: string
}

const defaultSettings: AppSettings = {
  title: import.meta.env.VITE_APP_TITLE,
  theme: '#409EFF',
  sideTheme: 'theme-dark',
  showSettings: true,
  navType: 1,
  tagsView: true,
  tagsViewPersist: false,
  tagsIcon: false,
  tagsViewStyle: 'card',
  fixedHeader: true,
  sidebarLogo: true,
  dynamicTitle: false,
  footerVisible: false,
  footerContent: 'Copyright © 2018-2026 RuoYi. All Rights Reserved.',
}

export default defaultSettings
