// ThemeBridge —— main.tsx 在 Provider 外无法读 Redux；
// 本组件在 Provider 内部读 settings.isDark/theme，返回配置好的 ConfigProvider（darkAlgorithm + colorPrimary）
// locale 也一并收进这里（原来在 main.tsx 的 ConfigProvider 上）

import { useEffect } from 'react'
import type { ReactNode } from 'react'
import { ConfigProvider, theme as antdTheme } from 'antd'
import zhCNModule from 'antd/locale/zh_CN'
import { useAppSelector } from '@/store/hooks'

// CJS interop 解包：当前预构建产物里 zhCN 是 { default: {...} } 双包装,
// 直接传给 ConfigProvider 会被当无效 locale 回落英文（Modal OK/Cancel、分页 10/page）。
// 任一层有 locale 字段即视为真 locale 对象。
const zhCN = ((zhCNModule as { default?: typeof zhCNModule }).default ?? zhCNModule) as typeof zhCNModule

export default function ThemeBridge({ children }: { children: ReactNode }) {
  const isDark = useAppSelector((t) => t.settings.isDark)
  const themeColor = useAppSelector((t) => t.settings.theme)

  // html.dark class 与 store 同步（启动时从持久化恢复 / toggleTheme 之外的路径变更也能跟上）
  useEffect(() => {
    document.documentElement.classList.toggle('dark', isDark)
  }, [isDark])

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
        token: { colorPrimary: themeColor },
      }}
    >
      {children}
    </ConfigProvider>
  )
}
