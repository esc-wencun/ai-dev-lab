// 占位页面集 —— 404/401/redirect 中转（登录/首页/锁屏已有真实或独立实现）
// redirect 中转：对位基准 views/redirect/index.vue（页签刷新机制依赖）

import { Button, Result } from 'antd'

export function NotFoundPage() {
  return (
    <Result
      status="404"
      title="404"
      subTitle="抱歉，您访问的页面不存在"
      extra={
        <Button type="primary" onClick={() => (window.location.href = '/index')}>
          返回首页
        </Button>
      }
    />
  )
}

export function UnauthorizedPage() {
  return (
    <Result
      status="403"
      title="401"
      subTitle="抱歉，你没有权限访问该页面"
      extra={
        <Button type="primary" onClick={() => (window.location.href = '/index')}>
          返回首页
        </Button>
      }
    />
  )
}
