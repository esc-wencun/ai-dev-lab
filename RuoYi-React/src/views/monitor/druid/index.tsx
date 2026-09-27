// 数据监控(druid) —— features.druidMonitor 降级页 + iFrame /druid/login.html

import PlatformGate from '@/components/PlatformGate'
import IFrame from '@/components/IFrame'

export default function Druid() {
  return (
    <PlatformGate feature="druidMonitor" subtitle="数据监控基于 Druid 连接池控制台，当前后端运行时未提供。">
      {() => <IFrame src={import.meta.env.VITE_APP_BASE_API + '/druid/login.html'} />}
    </PlatformGate>
  )
}
