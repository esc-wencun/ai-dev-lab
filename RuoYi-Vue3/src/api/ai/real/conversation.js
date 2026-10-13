import request from '@/utils/request'

// 会话 CRUD（字段与 Mock 逐字一致，走统一 request 实例——body code 语义/401/防重全复用）
export function listConversations() {
  return request({ url: '/ai/conversations', method: 'get' }).then(res => res.data)
}

export function createConversation({ title, modelId, agentId } = {}) {
  return request({ url: '/ai/conversations', method: 'post', data: { title, modelId, agentId } }).then(res => res.data)
}

export function renameConversation(conversationId, title) {
  return request({ url: `/ai/conversations/${conversationId}/title`, method: 'put', data: { title } }).then(res => res.data)
}

export function pinConversation(conversationId, pinned) {
  return request({ url: `/ai/conversations/${conversationId}/pin`, method: 'put', data: { pinned } }).then(res => res.data)
}

export function changeConversationModel(conversationId, modelId) {
  return request({ url: `/ai/conversations/${conversationId}/model`, method: 'put', data: { modelId } }).then(res => res.data)
}

/** 更新会话推理参数；-1 = 清除覆盖回落模型默认（后端 normalize 规则） */
export function updateConversationParams(conversationId, params) {
  return request({ url: `/ai/conversations/${conversationId}/params`, method: 'put', data: params }).then(res => res.data)
}

export function deleteConversation(conversationId) {
  return request({ url: `/ai/conversations/${conversationId}`, method: 'delete' })
}

export function listMessages(conversationId) {
  return request({ url: `/ai/conversations/${conversationId}/messages`, method: 'get' }).then(res => res.data)
}

export function listAgents() {
  return request({ url: '/ai/agents/options', method: 'get' }).then(res => res.data)
}

export function listChatModels() {
  return request({
    url: '/ai/model/list',
    method: 'get',
    params: { pageNum: 1, pageSize: 100, status: '0', modelType: 1 }
  }).then(res => res.rows)
}
