// 通用「未实现」占位页 —— 对位 deviations #2（build 砍除）与 gen 暂缓（用户 2026-09-28）
// 视觉与 features 降级页同构（antd Result）

import { Result } from 'antd'

interface NotImplementedProps {
  title?: string
  subTitle: string
}

export default function NotImplemented({ title = '该功能未实现', subTitle }: NotImplementedProps) {
  return (
    <div style={{ padding: 40 }}>
      <Result status="info" title={title} subTitle={subTitle} />
    </div>
  )
}
