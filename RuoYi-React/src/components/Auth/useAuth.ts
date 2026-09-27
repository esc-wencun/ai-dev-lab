// useAuth hook —— 脚本内权限判断（对位 $auth 插件）

import { useAppSelector } from '@/store/hooks'
import {
  hasPermi,
  hasPermiOr,
  hasPermiAnd,
  hasRole,
  hasRoleOr,
  hasRoleAnd,
} from '@/utils/permission'

export function useAuth() {
  const permissions = useAppSelector((s) => s.user.permissions)
  const roles = useAppSelector((s) => s.user.roles)
  return {
    hasPermi: (p: string) => hasPermi(permissions, p),
    hasPermiOr: (arr: string[]) => hasPermiOr(permissions, arr),
    hasPermiAnd: (arr: string[]) => hasPermiAnd(permissions, arr),
    hasRole: (r: string) => hasRole(roles, r),
    hasRoleOr: (arr: string[]) => hasRoleOr(roles, arr),
    hasRoleAnd: (arr: string[]) => hasRoleAnd(roles, arr),
  }
}
