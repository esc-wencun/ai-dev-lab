import request from '@/utils/request'

// 查询字典数据列表
export function listData(query?: Record<string, unknown>) {
  return request({ url: '/system/dict/data/list', method: 'get', params: query })
}

// 查询字典数据详细
export function getData(dictCode: number | string) {
  return request({ url: '/system/dict/data/' + dictCode, method: 'get' })
}

// 根据字典类型查询字典数据信息（useDict 消费的核心接口）
export function getDicts(dictType: string) {
  return request({ url: '/system/dict/data/type/' + dictType, method: 'get' })
}

// 新增字典数据
export function addData(data: Record<string, unknown>) {
  return request({ url: '/system/dict/data', method: 'post', data })
}

// 修改字典数据
export function updateData(data: Record<string, unknown>) {
  return request({ url: '/system/dict/data', method: 'put', data })
}

// 删除字典数据
export function delData(dictCode: number | string | (number | string)[]) {
  return request({ url: '/system/dict/data/' + dictCode, method: 'delete' })
}
