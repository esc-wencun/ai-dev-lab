import request from '@/utils/request'

// 查询AI厂商列表
export function listProvider(query) {
  return request({
    url: '/ai/provider/list',
    method: 'get',
    params: query
  })
}

// 全部正常厂商（下拉用）
export function selectAllProvider() {
  return request({
    url: '/ai/provider/selectAll',
    method: 'get'
  })
}

// 查询AI厂商明细
export function getProvider(providerId) {
  return request({
    url: '/ai/provider/' + providerId,
    method: 'get'
  })
}

// 新增AI厂商
export function addProvider(data) {
  return request({
    url: '/ai/provider',
    method: 'post',
    data: data
  })
}

// 修改AI厂商
export function updateProvider(data) {
  return request({
    url: '/ai/provider',
    method: 'put',
    data: data
  })
}

// 删除AI厂商
export function delProvider(providerIds) {
  return request({
    url: '/ai/provider/' + providerIds,
    method: 'delete'
  })
}
