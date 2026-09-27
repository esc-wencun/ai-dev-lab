import request from '@/utils/request'

// 查询菜单列表
export function listMenu(query?: Record<string, unknown>) {
  return request({ url: '/system/menu/list', method: 'get', params: query })
}

// 查询菜单详细
export function getMenu(menuId: number | string) {
  return request({ url: '/system/menu/' + menuId, method: 'get' })
}

// 查询菜单下拉树结构
export function treeselect() {
  return request({ url: '/system/menu/treeselect', method: 'get' })
}

// 根据角色ID查询菜单下拉树结构（含已勾选菜单 keys）
export function roleMenuTreeselect(roleId: number | string) {
  return request({ url: '/system/menu/roleMenuTreeselect/' + roleId, method: 'get' })
}

// 新增菜单
export function addMenu(data: Record<string, unknown>) {
  return request({ url: '/system/menu', method: 'post', data })
}

// 修改菜单
export function updateMenu(data: Record<string, unknown>) {
  return request({ url: '/system/menu', method: 'put', data })
}

// 保存菜单排序（批量：menuIds + orderNums）
export function updateMenuSort(data: Record<string, unknown>) {
  return request({ url: '/system/menu/updateSort', method: 'put', data })
}

// 删除菜单
export function delMenu(menuId: number | string | (number | string)[]) {
  return request({ url: '/system/menu/' + menuId, method: 'delete' })
}
