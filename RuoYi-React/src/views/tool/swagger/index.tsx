// 接口文档(swagger) —— features.swaggerDocs 降级页 + iFrame /swagger-ui/index.html

import PlatformGate from '@/components/PlatformGate'
import IFrame from '@/components/IFrame'

export default function Swagger() {
  return (
    <PlatformGate feature="swaggerDocs" subtitle="系统接口页基于 springdoc swagger-ui，当前后端运行时未提供。">
      {() => <IFrame src={import.meta.env.VITE_APP_BASE_API + '/swagger-ui/index.html'} />}
    </PlatformGate>
  )
}
