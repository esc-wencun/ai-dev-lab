import request from '@/utils/request'

// 查询AI模型列表
export function listModel(query) {
  return request({
    url: '/ai/model/list',
    method: 'get',
    params: query
  })
}

// 查询AI模型明细
export function getModel(modelId) {
  return request({
    url: '/ai/model/' + modelId,
    method: 'get'
  })
}

// 新增模型
export function addModel(data) {
  return request({
    url: '/ai/model',
    method: 'post',
    data: data
  })
}

// 修改模型
export function updateModel(data) {
  return request({
    url: '/ai/model',
    method: 'put',
    data: data
  })
}

// 删除模型
export function delModel(modelIds) {
  return request({
    url: '/ai/model/' + modelIds,
    method: 'delete'
  })
}
