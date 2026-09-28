// 权限判断纯函数矩阵 —— 语义对照 RuoYi-Vue3 src/plugins/auth.js + directive/permission
// 通配:权限 *:*:*、角色 admin;ROLE_DEFAULT 无特殊逻辑(与基准一致)
import { describe, expect, it } from 'vitest'
import {
  authPermission,
  authRole,
  hasPermi,
  hasPermiAnd,
  hasPermiOr,
  hasRole,
  hasRoleAnd,
  hasRoleOr,
} from './permission'

describe('authPermission / hasPermi', () => {
  it('精确命中', () => {
    expect(hasPermi(['system:user:list'], 'system:user:list')).toBe(true)
  })
  it('不命中 / 空持有列表', () => {
    expect(hasPermi(['system:user:list'], 'system:user:edit')).toBe(false)
    expect(hasPermi([], 'system:user:list')).toBe(false)
  })
  it('要求串含 *:*:* 直接放行(基准 hasPermi 特判)', () => {
    expect(hasPermi([], '*:*:*')).toBe(true)
  })
  it('持有 *:*:* 通配放行一切具体权限', () => {
    expect(authPermission(['*:*:*'], 'monitor:job:query')).toBe(true)
  })
  it('要求串为空 → false', () => {
    expect(authPermission(['system:user:list'], '')).toBe(false)
  })
})

describe('authRole / hasRole', () => {
  it('精确命中', () => {
    expect(hasRole(['common'], 'common')).toBe(true)
  })
  it('不命中 / 空持有列表', () => {
    expect(hasRole(['common'], 'audit')).toBe(false)
    expect(hasRole([], 'common')).toBe(false)
  })
  it('要求串含 admin 直接放行(基准 hasRole 特判)', () => {
    expect(hasRole([], 'admin')).toBe(true)
  })
  it('持有 admin 角色放行一切角色要求', () => {
    expect(authRole(['admin'], 'audit')).toBe(true)
  })
  it('ROLE_DEFAULT 无特殊逻辑(按普通串精确匹配)', () => {
    expect(hasRole(['ROLE_DEFAULT'], 'ROLE_DEFAULT')).toBe(true)
    expect(hasRole(['common'], 'ROLE_DEFAULT')).toBe(false)
  })
})

describe('hasPermiOr / hasPermiAnd', () => {
  const perms = ['system:user:list', 'monitor:job:query']
  it('Or:命中任一', () => {
    expect(hasPermiOr(perms, ['system:user:list', 'system:config:add'])).toBe(true)
    expect(hasPermiOr(perms, ['system:config:add'])).toBe(false)
  })
  it('Or:要求串含 *:*:* 全放行', () => {
    expect(hasPermiOr([], ['system:config:add', '*:*:*'])).toBe(true)
  })
  it('And:全部命中', () => {
    expect(hasPermiAnd(perms, perms)).toBe(true)
    expect(hasPermiAnd(perms, ['system:user:list', 'system:config:add'])).toBe(false)
  })
  it('And:持有 *:*:* 通配时全部放行', () => {
    expect(hasPermiAnd(['*:*:*'], ['a:b:c', 'd:e:f'])).toBe(true)
  })
  it('And:空数组 every → true(vanilla 语义,与基准一致)', () => {
    expect(hasPermiAnd(perms, [])).toBe(true)
  })
})

describe('hasRoleOr / hasRoleAnd', () => {
  const roles = ['common', 'ROLE_DEFAULT']
  it('Or:命中任一', () => {
    expect(hasRoleOr(roles, ['common', 'audit'])).toBe(true)
    expect(hasRoleOr(roles, ['audit'])).toBe(false)
  })
  it('Or:要求串含 admin 全放行', () => {
    expect(hasRoleOr([], ['audit', 'admin'])).toBe(true)
  })
  it('And:全部命中', () => {
    expect(hasRoleAnd(roles, roles)).toBe(true)
    expect(hasRoleAnd(roles, ['common', 'audit'])).toBe(false)
  })
  it('And:持有 admin 放行一切', () => {
    expect(hasRoleAnd(['admin'], ['common', 'audit'])).toBe(true)
  })
})
