// tagsView slice reducer 单测:直接构造 action 序列断言 state(不渲染组件、不发请求)
// 初始 state 用 tagsViewReducer(undefined, { type: '@@INIT' }) 获取(initialState 常量未导出)
// 模块无 DOM 依赖,走默认 node 环境
import { describe, expect, it } from 'vitest'
import tagsViewReducer, {
  addAffixView,
  addView,
  delAllViews,
  delLeftViews,
  delOthersViews,
  delRightViews,
  updateVisitedView,
} from './tagsView'
import type { TagView, TagsViewState } from './tagsView'

function init(): TagsViewState {
  return tagsViewReducer(undefined, { type: '@@INIT' }) as TagsViewState
}

function view(partial: Partial<TagView>): Partial<TagView> {
  return partial
}

describe('tagsView reducer', () => {
  it('初始 state:三个数组均为空', () => {
    const s = init()
    expect(s.visitedViews).toEqual([])
    expect(s.cachedViews).toEqual([])
    expect(s.iframeViews).toEqual([])
  })

  describe('addView', () => {
    it('新页签入 visitedViews,path 判重不重复 push', () => {
      let s = init()
      s = tagsViewReducer(s, addView(view({ path: '/system/user', name: 'User', title: '用户' })))
      s = tagsViewReducer(s, addView(view({ path: '/system/user', name: 'User', title: '用户' })))
      expect(s.visitedViews).toHaveLength(1)
      expect(s.visitedViews[0]).toMatchObject({
        path: '/system/user',
        fullPath: '/system/user',
        name: 'User',
        title: '用户',
      })
    })

    it("name 存在且 !noCache → 进 cachedViews(且缓存名不重复)", () => {
      let s = init()
      s = tagsViewReducer(s, addView(view({ path: '/a', name: 'A' })))
      s = tagsViewReducer(s, addView(view({ path: '/b', name: 'A' }))) // 同名再 push 应被 includes 挡住
      expect(s.cachedViews).toEqual(['A'])
    })

    it('noCache 不进 cachedViews', () => {
      let s = init()
      s = tagsViewReducer(s, addView(view({ path: '/a', name: 'A', meta: { noCache: true } })))
      expect(s.cachedViews).toEqual([])
      expect(s.visitedViews.map((v) => v.path)).toEqual(['/a'])
    })

    it('无名页签不进 cachedViews', () => {
      let s = init()
      s = tagsViewReducer(s, addView(view({ path: '/anon' })))
      expect(s.visitedViews).toHaveLength(1)
      expect(s.cachedViews).toEqual([])
    })

    it('meta.link → 同时进 iframeViews(按 path 判重)', () => {
      let s = init()
      s = tagsViewReducer(
        s,
        addView(view({ path: '/inner', name: 'Inner', title: '内嵌', meta: { link: 'https://a.com' } })),
      )
      s = tagsViewReducer(
        s,
        addView(view({ path: '/inner', name: 'Inner', title: '内嵌', meta: { link: 'https://a.com' } })),
      )
      expect(s.iframeViews).toHaveLength(1)
      expect(s.iframeViews[0]).toMatchObject({ path: '/inner', title: '内嵌' })
      expect(s.visitedViews).toHaveLength(1)
    })

    it('title 缺省补 no-name', () => {
      let s = init()
      s = tagsViewReducer(s, addView(view({ path: '/t' })))
      expect(s.visitedViews[0].title).toBe('no-name')
    })
  })

  describe('addAffixView', () => {
    it('unshift 置顶且判重', () => {
      let s = init()
      s = tagsViewReducer(s, addView(view({ path: '/user', name: 'User' })))
      s = tagsViewReducer(s, addAffixView(view({ path: '/index', name: 'Index', meta: { affix: true } })))
      s = tagsViewReducer(s, addAffixView(view({ path: '/index', name: 'Index', meta: { affix: true } })))
      expect(s.visitedViews.map((v) => v.path)).toEqual(['/index', '/user'])
    })
  })

  describe('delOthersViews', () => {
    it('保留 affix + 目标;iframeViews 只留目标;cachedViews 只留现存 visited 的 name', () => {
      let s = init()
      s = tagsViewReducer(s, addAffixView(view({ path: '/index', name: 'Index', meta: { affix: true } })))
      s = tagsViewReducer(s, addView(view({ path: '/a', name: 'A' })))
      s = tagsViewReducer(s, addView(view({ path: '/b', name: 'B' })))
      s = tagsViewReducer(s, addView(view({ path: '/iframe', name: 'I', meta: { link: 'https://x.com' } })))
      s = tagsViewReducer(s, delOthersViews(view({ path: '/a', name: 'A' })))
      expect(s.visitedViews.map((v) => v.path)).toEqual(['/index', '/a'])
      expect(s.iframeViews).toEqual([]) // 目标不是 iframe → iframeViews 清空
      // affix 经 addAffixView 进入,不写 cachedViews;B/I 的缓存名随页签消失
      expect(s.cachedViews).toEqual(['A'])
    })

    it('目标本身是 iframe 页签时保留在 iframeViews', () => {
      let s = init()
      s = tagsViewReducer(s, addView(view({ path: '/iframe', name: 'I', meta: { link: 'https://x.com' } })))
      s = tagsViewReducer(s, addView(view({ path: '/iframe2', name: 'I2', meta: { link: 'https://y.com' } })))
      s = tagsViewReducer(s, delOthersViews(view({ path: '/iframe', name: 'I' })))
      expect(s.iframeViews.map((v) => v.path)).toEqual(['/iframe'])
    })
  })

  describe('delAllViews', () => {
    it('只留 affix,iframeViews 清空,cachedViews 同步', () => {
      let s = init()
      s = tagsViewReducer(s, addAffixView(view({ path: '/index', name: 'Index', meta: { affix: true } })))
      s = tagsViewReducer(s, addView(view({ path: '/a', name: 'A' })))
      s = tagsViewReducer(s, addView(view({ path: '/b', name: 'B' })))
      s = tagsViewReducer(s, addView(view({ path: '/ifr', meta: { link: 'https://x.com' } })))
      s = tagsViewReducer(s, delAllViews())
      expect(s.visitedViews.map((v) => v.path)).toEqual(['/index'])
      expect(s.iframeViews).toEqual([])
      // affix 名不经 addView 写入 cachedViews,故全清后缓存为空
      expect(s.cachedViews).toEqual([])
    })
  })

  describe('delLeftViews / delRightViews', () => {
    function seeded(): TagsViewState {
      let s = init()
      s = tagsViewReducer(s, addAffixView(view({ path: '/index', name: 'Index', meta: { affix: true } })))
      s = tagsViewReducer(s, addView(view({ path: '/a', name: 'A' })))
      s = tagsViewReducer(s, addView(view({ path: '/b', name: 'B' })))
      s = tagsViewReducer(s, addView(view({ path: '/c', name: 'C' })))
      return s
    }

    it('delLeftViews:删目标左侧全部(affix 保留),cached 清孤儿', () => {
      const s = tagsViewReducer(seeded(), delLeftViews(view({ path: '/b', name: 'B' })))
      expect(s.visitedViews.map((v) => v.path)).toEqual(['/index', '/b', '/c'])
      expect(s.cachedViews).toEqual(['B', 'C']) // A 的缓存名随页签消失(Index 不在缓存,见 addAffixView 用例)
    })

    it('delRightViews:删目标右侧全部(affix 保留),cached 清孤儿', () => {
      const s = tagsViewReducer(seeded(), delRightViews(view({ path: '/b', name: 'B' })))
      expect(s.visitedViews.map((v) => v.path)).toEqual(['/index', '/a', '/b'])
      expect(s.cachedViews).toEqual(['A', 'B']) // C 的缓存名随页签消失
    })

    it('目标不存在(idx=-1)时 delLeftViews 现实现保留全部非 affix 项(i>=-1 恒真)', () => {
      const s = tagsViewReducer(seeded(), delLeftViews(view({ path: '/nope' })))
      expect(s.visitedViews.map((v) => v.path)).toEqual(['/index', '/a', '/b', '/c'])
    })

    it('affix 页签即使位于目标右侧也不被 delRightViews 删除,非 affix 右侧项被删', () => {
      let s = init()
      s = tagsViewReducer(s, addView(view({ path: '/b', name: 'B' })))
      s = tagsViewReducer(s, addView(view({ path: '/a', name: 'A' })))
      s = tagsViewReducer(s, addView(view({ path: '/affix', name: 'R', meta: { affix: true } })))
      s = tagsViewReducer(s, addView(view({ path: '/c', name: 'C' })))
      s = tagsViewReducer(s, delRightViews(view({ path: '/a', name: 'A' })))
      expect(s.visitedViews.map((v) => v.path)).toEqual(['/b', '/a', '/affix'])
      expect(s.cachedViews).toEqual(['B', 'A', 'R']) // C 的缓存名随页签清理
    })
  })

  describe('updateVisitedView', () => {
    it('按 path 合并字段', () => {
      let s = init()
      s = tagsViewReducer(s, addView(view({ path: '/a', name: 'A', title: '旧标题' })))
      s = tagsViewReducer(s, updateVisitedView(view({ path: '/a', name: 'A', title: '新标题', fullPath: '/a?x=1' })))
      expect(s.visitedViews[0]).toMatchObject({ path: '/a', title: '新标题', fullPath: '/a?x=1' })
      expect(s.visitedViews).toHaveLength(1)
    })

    it('path 不存在时无副作用', () => {
      let s = init()
      s = tagsViewReducer(s, addView(view({ path: '/a', name: 'A' })))
      const before = JSON.parse(JSON.stringify(s.visitedViews))
      s = tagsViewReducer(s, updateVisitedView(view({ path: '/nope', title: 'x' })))
      expect(s.visitedViews).toEqual(before)
    })
  })
})
