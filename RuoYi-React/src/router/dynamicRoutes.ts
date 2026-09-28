// 五条前端自有动态路由 —— 纯数据定义（对位 RuoYi-Vue3 src/router/index.js 的 dynamicRoutes）
// 独立成模块的原因：permission slice 需静态导入本数据做权限过滤；若从 routes.tsx 导入会形成
// permission → routes（含 Layout element）→ layout 组件链 → store 的模块循环，
// 首次求值入口为 App→permission 时触发 "Cannot access 'permission' before initialization"。
// 本文件只含类型与字符串，无任何 React element / 组件导入。

import type { RouteItem } from '@/store/modules/permission'

export interface DynamicRouteDef extends RouteItem {
  permissions?: string[]
  roles?: string[]
}

export const dynamicRoutes: DynamicRouteDef[] = [
  {
    path: '/system/user-auth',
    component: 'Layout',
    hidden: true,
    permissions: ['system:user:edit'],
    meta: {},
    children: [
      {
        path: 'role/:userId',
        component: 'system/user/authRole',
        name: 'AuthRole',
        meta: { title: '分配角色', activeMenu: '/system/user' },
      },
    ],
  },
]

// 五条中其余四条（与基准一一对应）
export const dynamicRoutesRest: DynamicRouteDef[] = [
  {
    path: '/system/role-auth',
    component: 'Layout',
    hidden: true,
    permissions: ['system:role:edit'],
    meta: {},
    children: [
      {
        path: 'user/:roleId',
        component: 'system/role/authUser',
        name: 'AuthUser',
        meta: { title: '分配用户', activeMenu: '/system/role' },
      },
    ],
  },
]

export const dynamicRoutesPart3: DynamicRouteDef[] = [
  {
    path: '/system/dict-data',
    component: 'Layout',
    hidden: true,
    permissions: ['system:dict:list'],
    meta: {},
    children: [
      {
        path: 'index/:dictId',
        component: 'system/dict/data',
        name: 'Data',
        meta: { title: '字典数据', activeMenu: '/system/dict' },
      },
    ],
  },
]

export const dynamicRoutesPart4: DynamicRouteDef[] = [
  {
    path: '/monitor/job-log',
    component: 'Layout',
    hidden: true,
    permissions: ['monitor:job:list'],
    meta: {},
    children: [
      {
        path: 'index/:jobId',
        component: 'monitor/job/log',
        name: 'JobLog',
        meta: { title: '调度日志', activeMenu: '/monitor/job' },
      },
    ],
  },
]

export const dynamicRoutesPart5: DynamicRouteDef[] = [
  {
    path: '/tool/gen-edit',
    component: 'Layout',
    hidden: true,
    permissions: ['tool:gen:edit'],
    meta: {},
    children: [
      {
        path: 'index/:tableId',
        component: 'tool/gen/editTable',
        name: 'GenEdit',
        meta: { title: '修改生成配置', activeMenu: '/tool/gen' },
      },
    ],
  },
]

// 合并导出：五条 = 首条 + 其余四条分片（分片写法规避超长写入）
export const allDynamicRoutes: DynamicRouteDef[] = [
  ...dynamicRoutes,
  ...dynamicRoutesRest,
  ...dynamicRoutesPart3,
  ...dynamicRoutesPart4,
  ...dynamicRoutesPart5,
]
