import request from '@/utils/request'

// 查询字典类型列表
export function listType(query?: Record<string, unknown>) {
  return request({ url: '/system/dict/type/list', method: 'get', params: query })
}

// 查询字典类型详细
export function getType(dictId: number | string) {
  return request({ url: '/system/dict/type/' + dictId, method: 'get' })
}

// 新增字典类型
export function addType(data: Record<string, unknown>) {
  return request({ url: '/system/dict/type', method: 'post', data })
}

// 修改字典类型
export function updateType(data: Record<string, unknown>) {
  return request({ url: '/system/dict/type', method: 'put', data })
}

// 删除字典类型
export function delType(dictId: number | string | (number | string)[]) {
  return request({ url: '/system/dict/type/' + dictId, method: 'delete' })
}

// 刷新字典缓存
export function refreshCache() {
  return request({ url: '/system/dict/type/refreshCache', method: 'delete' })
}

// 获取字典选择框列表
export function optionselect() {
  return request({ url: '/system/dict/type/optionselect', method: 'get' })
}
