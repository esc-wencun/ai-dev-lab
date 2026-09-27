import request from '@/utils/request'

// 查询部门列表
export function listDept(query?: Record<string, unknown>) {
  return request({ url: '/system/dept/list', method: 'get', params: query })
}

// 查询部门列表（排除自身及子节点——新增/修改上级部门选择用）
export function listDeptExcludeChild(deptId: number | string) {
  return request({ url: '/system/dept/list/exclude/' + deptId, method: 'get' })
}

// 查询部门详细
export function getDept(deptId: number | string) {
  return request({ url: '/system/dept/' + deptId, method: 'get' })
}

// 新增部门
export function addDept(data: Record<string, unknown>) {
  return request({ url: '/system/dept', method: 'post', data })
}

// 修改部门
export function updateDept(data: Record<string, unknown>) {
  return request({ url: '/system/dept', method: 'put', data })
}

// 保存部门排序（批量：deptIds + orderNums）
export function updateDeptSort(data: Record<string, unknown>) {
  return request({ url: '/system/dept/updateSort', method: 'put', data })
}

// 删除部门
export function delDept(deptId: number | string | (number | string)[]) {
  return request({ url: '/system/dept/' + deptId, method: 'delete' })
}
