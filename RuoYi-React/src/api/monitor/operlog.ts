import request from '@/utils/request'

// 查询操作日志列表
export function list(query?: Record<string, unknown>) {
  return request({ url: '/monitor/operlog/list', method: 'get', params: query })
}

// 删除操作日志
export function delOperlog(operId: number | string | (number | string)[]) {
  return request({ url: '/monitor/operlog/' + operId, method: 'delete' })
}

// 清空操作日志
export function cleanOperlog() {
  return request({ url: '/monitor/operlog/clean', method: 'delete' })
}
