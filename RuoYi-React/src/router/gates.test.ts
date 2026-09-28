// @vitest-environment jsdom
// gates 纯逻辑单测:deriveRouteName + buildRouteObjects 拍平行为
// makeViewElement / import.meta.glob 命中分支不测(依赖 Vite 虚拟模块/懒加载),
// element 只断言两种可判定形态:Outlet 引用(component 为 'Layout'/缺省)与占位组件(glob 未命中)
import { describe, expect, it } from 'vitest'
import { isValidElement } from 'react'
import type { ReactElement } from 'react'
import { Outlet } from 'react-router'
import { buildRouteObjects, deriveRouteName } from './gates'
import type { RouteItem } from '@/store/modules/permission'

describe('deriveRouteName', () => {
  it('根路径与空路径 → Index', () => {
    expect(deriveRouteName('/')).toBe('Index')
    expect(deriveRouteName('')).toBe('Index')
  })

  it('普通路径 PascalCase(仅段首字母大写)', () => {
    expect(deriveRouteName('/system/user')).toBe('SystemUser')
    expect(deriveRouteName('/monitor/operlog')).toBe('MonitorOperlog')
  })

  it('动态段剔除(:后整段丢弃)', () => {
    expect(deriveRouteName('/system/user/role/:userId')).toBe('SystemUserRole')
    expect(deriveRouteName('/user/profile/:activeTab?')).toBe('UserProfile')
  })

  it('真实 dynamicRoutes 形状(正则约束 + 连字符仅首字母大写)', () => {
    expect(deriveRouteName('/system/user-auth/role/:userId(d+)')).toBe('SystemUser-authRole')
  })

  it('尾斜杠不产生空段', () => {
    expect(deriveRouteName('/index/')).toBe('Index')
  })
})

// 后端菜单树样例:覆盖 唯一可见子级 / ParentView 多级拍平 / hidden / http(s) 外链 / 顶层无子级
function sampleMenus(): RouteItem[] {
  return [
    {
      path: '/system',
      component: 'Layout',
      alwaysShow: true,
      meta: { icon: 'system', title: '系统管理' },
      children: [
        {
          path: 'user',
          component: 'system/user/index',
          name: 'User',
          meta: { title: '用户管理', icon: 'user', noCache: true, affix: true, activeMenu: '/index' },
        },
        {
          path: 'role',
          component: 'ParentView',
          meta: { title: '角色管理' },
          children: [
            { path: 'auth', component: 'system/role/index', name: 'Auth', meta: { title: '分配角色' } },
          ],
        },
        { path: 'bare', meta: { title: '裸页面' } }, // 无 component → Outlet
        { path: 'ghost', component: 'no/such/view', meta: { title: '占位页' } }, // glob 未命中 → 占位 element
        { path: 'hidden-x', component: 'system/x/index', hidden: true, meta: { title: '隐藏页' } },
      ],
    },
    {
      path: '/deep',
      component: 'Layout',
      meta: { title: '多级' },
      children: [
        {
          path: 'mid',
          component: 'ParentView',
          meta: {},
          children: [
            {
              path: 'inner',
              component: 'ParentView',
              meta: {},
              children: [
                { path: 'leaf', component: 'tool/gen/index', meta: { title: '深层叶子' } },
              ],
            },
          ],
        },
      ],
    },
    { path: '/about', component: 'Layout', meta: { title: '关于' } }, // 顶层无子级
    { path: '/hidden-top', component: 'Layout', hidden: true, meta: {} }, // 顶层 hidden 跳过
    { path: '/ext', component: 'Layout', meta: { title: '外链', link: 'https://github.com' } }, // 顶层外链跳过
    {
      path: '/leaf-link',
      component: 'Layout',
      meta: {},
      children: [
        // 子级 http 外链:不注册 Route(侧边栏渲染 <a>),不产生叶子
        { path: 'https://rust.org', component: 'x', meta: { title: '子级外链', link: 'https://rust.org' } },
      ],
    },
  ]
}

describe('buildRouteObjects', () => {
  const routes = buildRouteObjects(sampleMenus())

  it('多级嵌套叶子全部收集到 Layout 层,hidden/外链跳过', () => {
    expect(routes.map((r) => r.path)).toEqual([
      '/system/user',
      '/system/role/auth', // 唯一可见子级直接取子路径(拼父路径)
      '/system/bare',
      '/system/ghost',
      '/deep/mid/inner/leaf', // ParentView 中间层递归拍平,子路径逐级拼父路径
      '/about',
    ])
  })

  it('handle 携带 name/title/noCache/affix/icon/activeMenu/meta/metaLink', () => {
    const user = routes.find((r) => r.path === '/system/user')!
    expect(user.handle).toEqual({
      name: 'User', // 显式 name 优先
      title: '用户管理',
      meta: { title: '用户管理', icon: 'user', noCache: true, affix: true, activeMenu: '/index' },
      metaLink: null,
      activeMenu: '/index',
      noCache: true,
      affix: true,
      icon: 'user',
    })
  })

  it('未显式命名的叶子按拍平后全路径派生 name', () => {
    const bare = routes.find((r) => r.path === '/system/bare')!
    expect(bare.handle?.name).toBe('SystemBare')
    const deep = routes.find((r) => r.path === '/deep/mid/inner/leaf')!
    expect(deep.handle?.name).toBe('DeepMidInnerLeaf')
    const about = routes.find((r) => r.path === '/about')!
    expect(about.handle?.name).toBe('About')
  })

  it("component 为 'Layout'/缺省 → Outlet;不存在路径 → 占位 element", () => {
    const about = routes.find((r) => r.path === '/about')!
    expect(isValidElement(about.element)).toBe(true)
    expect((about.element as ReactElement).type).toBe(Outlet)

    const bare = routes.find((r) => r.path === '/system/bare')!
    expect((bare.element as ReactElement).type).toBe(Outlet)

    const ghost = routes.find((r) => r.path === '/system/ghost')!
    expect(isValidElement(ghost.element)).toBe(true)
    expect((ghost.element as ReactElement).type).not.toBe(Outlet)
    expect(((ghost.element as ReactElement).type as unknown as { name?: string }).name).toBe(
      'ViewPlaceholder',
    )
  })

  it('顶层 hidden 与 http(s) 外链(meta.link)不产生路由', () => {
    expect(routes.some((r) => r.path.startsWith('/hidden-top'))).toBe(false)
    expect(routes.some((r) => r.path === '/ext')).toBe(false)
    expect(routes.some((r) => r.path.includes('rust.org'))).toBe(false)
  })
})
