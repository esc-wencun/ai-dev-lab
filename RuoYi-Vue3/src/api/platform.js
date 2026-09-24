import request from '@/utils/request'

// 获取平台信息（语言/框架版本与功能开关，数据监控等服务端特有页面据此降级提示）
export function getPlatformInfo() {
  return request({
    url: '/getPlatformInfo',
    method: 'get'
  })
}
