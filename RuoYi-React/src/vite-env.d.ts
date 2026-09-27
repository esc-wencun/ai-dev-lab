/// <reference types="vite/client" />

// vite-plugin-svg-icons 虚拟模块声明
declare module 'virtual:svg-icons-register'

interface ImportMetaEnv {
  /** 页面标题（document.title / Logo 文案） */
  readonly VITE_APP_TITLE: string
  /** 环境标识：development / staging / production */
  readonly VITE_APP_ENV: 'development' | 'staging' | 'production'
  /** axios baseURL：/dev-api | /stage-api | /prod-api */
  readonly VITE_APP_BASE_API: string
  /** 生产构建压缩方式（预留，当前未接压缩插件） */
  readonly VITE_BUILD_COMPRESS?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
