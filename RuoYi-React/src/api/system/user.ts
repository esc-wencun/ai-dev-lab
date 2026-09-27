import request from '@/utils/request'
import { parseStrEmpty } from '@/utils/ruoyi'

// 查询用户列表
export function listUser(query?: Record<string, unknown>) {
  return request({
    url: '/system/user/list',
    method: 'get',
    params: query,
  })
}

// 查询用户详细（不传 userId 时调尾斜杠端点拿角色岗位选项——与基准一致）
export function getUser(userId?: number | string) {
  return request({
    url: '/system/user/' + parseStrEmpty(String(userId ?? '')),
    method: 'get',
  })
}

// 新增用户
export function addUser(data: Record<string, unknown>) {
  return request({
    url: '/system/user',
    method: 'post',
    data,
  })
}

// 修改用户
export function updateUser(data: Record<string, unknown>) {
  return request({
    url: '/system/user',
    method: 'put',
    data,
  })
}

// 删除用户
export function delUser(userId: number | string | (number | string)[]) {
  return request({
    url: '/system/user/' + userId,
    method: 'delete',
  })
}

// 用户密码重置
export function resetUserPwd(userId: number | string, password: string) {
  const data = { userId, password }
  return request({
    url: '/system/user/resetPwd',
    method: 'put',
    data,
  })
}

// 用户状态修改
export function changeUserStatus(userId: number | string, status: string) {
  const data = { userId, status }
  return request({
    url: '/system/user/changeStatus',
    method: 'put',
    data,
  })
}

// 查询用户个人信息
export function getUserProfile() {
  return request({
    url: '/system/user/profile',
    method: 'get',
  })
}

// 修改用户个人信息
export function updateUserProfile(data: Record<string, unknown>) {
  return request({
    url: '/system/user/profile',
    method: 'put',
    data,
  })
}

// 用户密码重置（个人中心）
export function updateUserPwd(oldPassword: string, newPassword: string) {
  const data = { oldPassword, newPassword }
  return request({
    url: '/system/user/profile/updatePwd',
    method: 'put',
    data,
  })
}

// 用户头像上传（multipart，字段名 avatarfile 是三版后端解析契约）
// 注：基准 api 声明 urlencoded 是历史瑕疵，FormData 实际按 multipart 发送；React 版直接按 multipart 声明（deviations #12）
export function uploadAvatar(data: FormData) {
  return request({
    url: '/system/user/profile/avatar',
    method: 'post',
    headers: { 'Content-Type': 'multipart/form-data' },
    data,
  })
}

// 查询授权角色
export function getAuthRole(userId: number | string) {
  return request({
    url: '/system/user/authRole/' + userId,
    method: 'get',
  })
}

// 保存授权角色
export function updateAuthRole(data: Record<string, unknown>) {
  return request({
    url: '/system/user/authRole',
    method: 'put',
    params: data,
  })
}

// 查询部门下拉树结构
export function deptTreeSelect() {
  return request({
    url: '/system/user/deptTree',
    method: 'get',
  })
}
