// @vitest-environment jsdom
// filterDynamicRoutes 纯函数单测(permissions → hasPermiOr / roles → hasRoleOr / 两者都无 → 丢弃)
// 语义对照 src/utils/permission.ts:权限通配 *:*:*,角色通配 admin
import { describe, expect, it } from 'vitest'
import { filterDynamicRoutes } from './permission'
import type { RouteItem } from './permission'

type DynRoute = RouteItem & { permissions?: string[]; roles?: string[] }

const routes: DynRoute[] = [
  { path: '/a', permissions: ['system:user:edit'], meta: {} },
  { path: '/b', permissions: ['monitor:job:list', 'system:config:add'], meta: {} },
  { path: '/c', roles: ['common'], meta: {} },
  { path: '/d', roles: ['audit'], meta: {} },
  { path: '/e', meta: {} }, // 无 permissions 无 roles → 必丢弃
]

describe('filterDynamicRoutes', () => {
  it('permissions 命中(或语义:命中任一即可)', () => {
    const out = filterDynamicRoutes(routes, ['system:user:edit', 'system:user:query'], [])
    expect(out.map((r) => r.path)).toEqual(['/a'])
    // 持有第二组中的任一权限即通过
    const out2 = filterDynamicRoutes(routes, ['system:config:add'], [])
    expect(out2.map((r) => r.path)).toEqual(['/b'])
  })

  it('permissions 不匹配 → 丢弃', () => {
    const out = filterDynamicRoutes(routes, ['system:role:list'], [])
    expect(out.map((r) => r.path)).toEqual([])
  })

  it('用户权限含 *:*:* 通配 → 所有 permissions 路由全放行', () => {
    const out = filterDynamicRoutes(routes, ['*:*:*'], [])
    expect(out.map((r) => r.path)).toEqual(['/a', '/b'])
  })

  it('roles 命中:用户持 admin 角色放行一切 roles 路由(基准 admin 特判)', () => {
    const out = filterDynamicRoutes(routes, [], ['admin'])
    expect(out.map((r) => r.path)).toEqual(['/c', '/d'])
  })

  it('roles 普通匹配:精确命中任一即可', () => {
    const out = filterDynamicRoutes(routes, [], ['common', 'other'])
    expect(out.map((r) => r.path)).toEqual(['/c'])
  })

  it('permissions 与 roles 同时不满足的分支互不越界:permissions 路由不受角色放行,反之亦然', () => {
    const out = filterDynamicRoutes(routes, [], ['common'])
    expect(out.map((r) => r.path)).toEqual(['/c'])
    const out2 = filterDynamicRoutes(routes, ['system:user:edit'], [])
    expect(out2.some((r) => r.path === '/c')).toBe(false)
  })

  it('无 permissions 且无 roles 的路由一律丢弃(即使持管理员权限)', () => {
    const out = filterDynamicRoutes(
      [{ path: '/x', meta: {} }],
      ['*:*:*'],
      ['admin'],
    )
    expect(out).toEqual([])
  })

  it('空入参返回空', () => {
    expect(filterDynamicRoutes([], ['system:user:edit'], ['admin'])).toEqual([])
  })
})
