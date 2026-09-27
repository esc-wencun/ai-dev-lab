import request from '@/utils/request'

// 查询公告列表
export function listNotice(query?: Record<string, unknown>) {
  return request({ url: '/system/notice/list', method: 'get', params: query })
}

// 查询公告详细
export function getNotice(noticeId: number | string) {
  return request({ url: '/system/notice/' + noticeId, method: 'get' })
}

// 新增公告
export function addNotice(data: Record<string, unknown>) {
  return request({ url: '/system/notice', method: 'post', data })
}

// 修改公告
export function updateNotice(data: Record<string, unknown>) {
  return request({ url: '/system/notice', method: 'put', data })
}

// 删除公告
export function delNotice(noticeId: number | string | (number | string)[]) {
  return request({ url: '/system/notice/' + noticeId, method: 'delete' })
}

// 首页顶部公告列表（带已读状态）
export function listNoticeTop() {
  return request({ url: '/system/notice/listTop', method: 'get' })
}

// 标记公告已读
export function markNoticeRead(noticeId: number | string) {
  return request({ url: '/system/notice/markRead', method: 'post', params: { noticeId } })
}

// 批量标记已读
export function markNoticeReadAll(ids: string) {
  return request({ url: '/system/notice/markReadAll', method: 'post', params: { ids } })
}

// 查询公告已读用户列表
export function listNoticeReadUsers(query?: Record<string, unknown>) {
  return request({ url: '/system/notice/readUsers/list', method: 'get', params: query })
}
