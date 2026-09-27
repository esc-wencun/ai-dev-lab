// 权限判断 —— 对位 RuoYi-Vue3 src/plugins/auth.js + src/directive/permission/*
// 纯函数形态（供 Auth 组件 / 路由过滤 / useAuth hook 共用）
// 通配：权限 `*:*:*`、角色 `admin`；ROLE_DEFAULT 无特殊逻辑（与基准一致）

export function authPermission(permissions: string[], permission: string): boolean {
  if (permission.includes('*:*:*')) return true
  if (!permission || permission.length === 0) return false
  return permissions.some((p) => p === permission || p === '*:*:*')
}

export function authRole(roles: string[], role: string): boolean {
  if (role.includes('admin')) return true
  if (!role || role.length === 0) return false
  return roles.some((r) => r === role || r === 'admin')
}

export function hasPermi(permissions: string[], permission: string): boolean {
  return authPermission(permissions, permission)
}

export function hasPermiOr(permissions: string[], arr: string[]): boolean {
  return arr.some((p) => authPermission(permissions, p))
}

export function hasPermiAnd(permissions: string[], arr: string[]): boolean {
  return arr.every((p) => authPermission(permissions, p))
}

export function hasRole(roles: string[], role: string): boolean {
  return authRole(roles, role)
}

export function hasRoleOr(roles: string[], arr: string[]): boolean {
  return arr.some((r) => authRole(roles, r))
}

export function hasRoleAnd(roles: string[], arr: string[]): boolean {
  return arr.every((r) => authRole(roles, r))
}
