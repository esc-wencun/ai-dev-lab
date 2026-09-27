// PlatformGate —— features 降级组件（对位基准三页统一模式）
// 初值 true（接口慢/失败不误显降级页）→ getPlatformInfo 覆盖 → false 渲染降级提示 + 重新检测

import { useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import { Button, Result } from 'antd'
import { getPlatformInfo } from '@/api/platform'

interface PlatformGateProps {
  feature: 'druidMonitor' | 'serverMonitor' | 'swaggerDocs'
  subtitle: string
  children: (enabled: boolean) => ReactNode
}

export default function PlatformGate({ feature, subtitle, children }: PlatformGateProps) {
  const [features, setFeatures] = useState<Record<string, boolean>>({ [feature]: true } as Record<string, boolean>)

  const check = () => {
    getPlatformInfo().then((res) => {
      const data = res as unknown as { features?: Record<string, boolean> }
      setFeatures(data.features || {})
    })
  }

  useEffect(() => {
    check()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 与基准一致：只有明确 false 才降级；重新检测按钮覆盖 features
  if (features[feature] === false) {
    return (
      <div style={{ padding: 40 }}>
        <Result
          status="info"
          title="该功能仅 Java 版提供"
          subTitle={subtitle}
          extra={<Button type="primary" onClick={check}>重新检测</Button>}
        />
      </div>
    )
  }
  return <>{children(features[feature] !== false as boolean)}</>
}
