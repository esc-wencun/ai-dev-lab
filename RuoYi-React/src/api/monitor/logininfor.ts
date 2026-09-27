import request from '@/utils/request'

// 查询登录日志列表
export function list(query?: Record<string, unknown>) {
  return request({ url: '/monitor/logininfor/list', method: 'get', params: query })
}

// 删除登录日志
export function delLogininfor(infoId: number | string | (number | string)[]) {
  return request({ url: '/monitor/logininfor/' + infoId, method: 'delete' })
}

// 解锁用户登录状态
export function unlockLogininfor(userName: string) {
  return request({ url: '/monitor/logininfor/unlock/' + userName, method: 'get' })
}

// 清空登录日志
export function cleanLogininfor() {
  return request({ url: '/monitor/logininfor/clean', method: 'delete' })
}
