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

// 连通性测试（specs/2.0.3）：发送一次最小请求验证 厂商地址/密钥/模型 全链路
export function testModel(modelId) {
  return request({
    url: `/ai/model/${modelId}/test`,
    method: 'post',
    timeout: 60000
  })
}
