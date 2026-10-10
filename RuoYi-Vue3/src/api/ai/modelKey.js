import request from '@/utils/request'

// 查询AI密钥列表
export function listKey(query) {
  return request({
    url: '/ai/key/list',
    method: 'get',
    params: query
  })
}

// 按厂商查启用密钥（下拉用）
export function listKeyByProvider(providerId) {
  return request({
    url: '/ai/key/listByProvider/' + providerId,
    method: 'get'
  })
}

// 查询AI密钥明细
export function getKey(keyId) {
  return request({
    url: '/ai/key/' + keyId,
    method: 'get'
  })
}

// 新增密钥
export function addKey(data) {
  return request({
    url: '/ai/key',
    method: 'post',
    data: data
  })
}

// 修改密钥
export function updateKey(data) {
  return request({
    url: '/ai/key',
    method: 'put',
    data: data
  })
}

// 删除密钥
export function delKey(keyIds) {
  return request({
    url: '/ai/key/' + keyIds,
    method: 'delete'
  })
}
