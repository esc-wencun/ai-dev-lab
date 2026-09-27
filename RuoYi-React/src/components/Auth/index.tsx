// Auth 权限组件 —— 对位 v-hasPermi/v-hasRole 指令（deviations #8）
// 无权限渲染 null（等价基准的 DOM 移除）；通配：*:*:* / admin

import type { ReactNode } from 'react'
import { useAppSelector } from '@/store/hooks'
import { hasPermiOr, hasRoleOr } from '@/utils/permission'

interface AuthProps {
  children: ReactNode
  permissions?: string[]
  roles?: string[]
}

export default function Auth({ children, permissions, roles }: AuthProps) {
  const userPerms = useAppSelector((s) => s.user.permissions)
  const userRoles = useAppSelector((s) => s.user.roles)

  if (permissions && permissions.length > 0) {
    if (!hasPermiOr(userPerms, permissions)) return null
  }
  if (roles && roles.length > 0) {
    if (!hasRoleOr(userRoles, roles)) return null
  }
  return <>{children}</>
}
