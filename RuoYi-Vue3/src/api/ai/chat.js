import request from '@/utils/request'

// AI 对话冒烟（specs/2.0.0/2.0.1；2.1.0 正式会话接口就位后替换）
export function sendAiChat(data) {
  return request({
    url: '/ai/chat',
    method: 'post',
    data: data
  })
}
