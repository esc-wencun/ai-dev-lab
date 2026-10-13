import request from '@/utils/request'

// 智能体管理页（specs/2.0.4）：对齐 /ai/agents 五端点；suggestedQuestions 前端负责 入口序列化/出口已由后端转数组
export function listAgent(query) {
  return request({ url: '/ai/agents', method: 'get', params: query }).then(res => res.data)
}

export function getAgent(agentId) {
  return request({ url: `/ai/agents/${agentId}`, method: 'get' }).then(res => res.data)
}

export function addAgent(data) {
  return request({ url: '/ai/agents', method: 'post', data })
}

export function updateAgent(data) {
  return request({ url: '/ai/agents', method: 'put', data })
}

export function delAgent(agentId) {
  return request({ url: `/ai/agents/${agentId}`, method: 'delete' })
}
