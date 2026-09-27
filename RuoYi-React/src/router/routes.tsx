// 静态路由表 —— 对位 RuoYi-Vue3 src/router/index.js
// constantRoutes（公开）+ dynamicRoutes（前端自有，按 permissions 过滤的五条子页）
// 页面组件暂以占位/已实现页面混合，随 3.0.0+ 渐次补齐真实页面

import type { RouteObjectLite, RouteHandle } from './gates'
import { makeViewElement } from './gates'
import Layout from '@/layout'
import { createElement } from 'react'
import type { RouteItem } from '@/store/modules/permission'

export const constantRoutes: RouteObjectLite[] = [
  { path: '/login', element: makeViewElement('login') },
  { path: '/register', element: makeViewElement('register') },
  { path: '/lock', element: makeViewElement('lock') },
]

// 布局壳路由（登录后的一切页面挂这里）
export const layoutRoute: RouteObjectLite = {
  path: '/',
  element: createElementLayout(),
  children: [
    indexChild(),
    redirectChild(),
    profileChild(),
  ],
}

function createElementLayout() {
  return createElement(Layout)
}

function indexChild(): RouteObjectLite {
  return {
    path: '/index',
    element: makeViewElement('index'),
    handle: {
      name: 'Index',
      title: '首页',
      affix: true,
      icon: 'dashboard',
    } as RouteHandle,
  }
}

function redirectChild(): RouteObjectLite {
  return {
    path: '/redirect/:path(.*)',
    element: makeViewElement('redirect/index'),
  }
}

function profileChild(): RouteObjectLite {
  return {
    path: '/user/profile/:activeTab?',
    element: makeViewElement('system/user/profile/index'),
    handle: {
      name: 'Profile',
      title: '个人中心',
      icon: 'user',
    } as RouteHandle,
  }
}

// 五条前端自有动态路由（对位 dynamicRoutes，全部 hidden + 权限过滤）
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
        path: 'role/:userId(\d+)',
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
        path: 'user/:roleId(\d+)',
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
        path: 'index/:dictId(\d+)',
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
        path: 'index/:jobId(\d+)',
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
        path: 'index/:tableId(\d+)',
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

// 404 置底（Gate 组装时放最后，防动态路由未就绪误伤深层链接）
export const notFoundRoute: RouteObjectLite = {
  path: '*',
  element: makeViewElement('error/404'),
}

export const error401Route: RouteObjectLite = {
  path: '/401',
  element: makeViewElement('error/401'),
}
