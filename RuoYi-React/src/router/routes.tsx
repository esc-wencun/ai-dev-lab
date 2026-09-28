// 静态路由表 —— 对位 RuoYi-Vue3 src/router/index.js
// constantRoutes（公开）+ dynamicRoutes（前端自有，按 permissions 过滤的五条子页）
// 页面组件暂以占位/已实现页面混合，随 3.0.0+ 渐次补齐真实页面

import type { RouteObjectLite, RouteHandle } from './gates'
import { makeViewElement } from './gates'
import Layout from '@/layout'
import { createElement } from 'react'
// 五条前端自有动态路由的纯数据定义在 ./dynamicRoutes（独立模块切断 permission→routes→layout
// 的循环导入），此处 re-export 保持既有导入路径兼容
import {
  dynamicRoutes,
  dynamicRoutesRest,
  dynamicRoutesPart3,
  dynamicRoutesPart4,
  dynamicRoutesPart5,
  allDynamicRoutes,
} from './dynamicRoutes'
import type { DynamicRouteDef } from './dynamicRoutes'

export {
  dynamicRoutes,
  dynamicRoutesRest,
  dynamicRoutesPart3,
  dynamicRoutesPart4,
  dynamicRoutesPart5,
  allDynamicRoutes,
}
export type { DynamicRouteDef }

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
    // react-router v7 通配段语法是 `*`（v6 的 `:path(.*)` regex 写法在 v7 已移除，
    // 会导致 /redirect/xxx 匹配 404，页签刷新机制失效）
    path: '/redirect/*',
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

// 404 置底（Gate 组装时放最后，防动态路由未就绪误伤深层链接）
export const notFoundRoute: RouteObjectLite = {
  path: '*',
  element: makeViewElement('error/404'),
}

export const error401Route: RouteObjectLite = {
  path: '/401',
  element: makeViewElement('error/401'),
}
