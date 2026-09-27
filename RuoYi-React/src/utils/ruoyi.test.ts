import { describe, expect, it } from 'vitest'
import {
  parseTime,
  addDateRange,
  selectDictLabel,
  handleTree,
  tansParams,
  getNormalPath,
  blobValidate,
  parseStrEmpty,
} from './ruoyi'

// 全部断言对照 RuoYi-Vue3 src/utils/ruoyi.js 原实现行为
describe('parseTime', () => {
  it('Date 对象 → 默认格式', () => {
    const d = new Date(2026, 8, 27, 14, 30, 45)
    expect(parseTime(d)).toBe('2026-09-27 14:30:45')
  })

  it('秒级时间戳自动 ×1000', () => {
    const d = new Date(2026, 8, 27, 0, 0, 0)
    const sec = Math.floor(d.getTime() / 1000)
    expect(parseTime(sec)).toBe('2026-09-27 00:00:00')
  })

  it('ISO 字符串兼容（- / T / .SSS）', () => {
    expect(parseTime('2026-09-27T14:30:45.000')).toBe('2026-09-27 14:30:45')
  })

  it('自定义格式 + {a} 星期（2026-09-27 是周日）', () => {
    const d = new Date(2026, 8, 27, 9, 5, 3)
    expect(parseTime(d, '{y}-{m}-{d} {h}:{i}:{s} 周{a}')).toBe('2026-09-27 09:05:03 周日')
  })

  it('空值返回 null', () => {
    expect(parseTime(undefined)).toBeNull()
    expect(parseTime('')).toBeNull()
  })
})

describe('tansParams', () => {
  it('扁平参数拼接（尾带 &，与基准一致）', () => {
    expect(tansParams({ pageNum: 1, name: 'ruoyi' })).toBe('pageNum=1&name=ruoyi&')
  })

  it('跳过 null/空串/undefined，保留 0 和 false', () => {
    expect(tansParams({ a: null, b: '', c: undefined, d: 0, e: false })).toBe('d=0&e=false&')
  })

  it('嵌套对象展开为 prop[key]', () => {
    const out = tansParams({ params: { beginTime: '2026-09-27', endTime: '2026-09-28' } })
    expect(out).toBe('params%5BbeginTime%5D=2026-09-27&params%5BendTime%5D=2026-09-28&')
  })

  it('嵌套对象内部空值跳过', () => {
    const out = tansParams({ params: { beginTime: 'x', endTime: null } })
    expect(out).toBe('params%5BbeginTime%5D=x&')
  })

  it('中文编码', () => {
    const out = tansParams({ name: '系统管理' })
    expect(out).toBe('name=' + encodeURIComponent('系统管理') + '&')
  })
})

describe('addDateRange', () => {
  it('默认写入 beginTime/endTime', () => {
    const p = addDateRange({}, ['2026-09-27', '2026-09-28'])
    expect(p.params).toEqual({ beginTime: '2026-09-27', endTime: '2026-09-28' })
  })

  it('propName 后缀写入 beginXxx/endXxx', () => {
    const p = addDateRange({}, ['a', 'b'], 'UpTime')
    expect(p.params).toEqual({ beginUpTime: 'a', endUpTime: 'b' })
  })
})

describe('selectDictLabel', () => {
  const datas = [
    { label: '正常', value: '0' },
    { label: '停用', value: '1' },
  ]
  it('命中返回 label', () => {
    expect(selectDictLabel(datas, '0')).toBe('正常')
  })
  it('未命中返回原值', () => {
    expect(selectDictLabel(datas, '9')).toBe('9')
  })
  it('undefined 返回空串', () => {
    expect(selectDictLabel(datas, undefined)).toBe('')
  })
})

describe('handleTree', () => {
  it('平铺数据构造树', () => {
    const tree = handleTree<{ id: number; parentId: number; children?: unknown[] }>([
      { id: 1, parentId: 0 },
      { id: 2, parentId: 1 },
      { id: 3, parentId: 1 },
    ])
    expect(tree).toHaveLength(1)
    expect(tree[0].children).toHaveLength(2)
  })
})

describe('getNormalPath', () => {
  it('首个双斜杠合并、去尾斜杠', () => {
    expect(getNormalPath('//user/profile/')).toBe('/user/profile')
  })
})

describe('blobValidate', () => {
  it('JSON 错误体返回 false', () => {
    expect(blobValidate(new Blob(['{}'], { type: 'application/json' }))).toBe(false)
  })
  it('文件 blob 返回 true', () => {
    expect(blobValidate(new Blob(['x'], { type: 'application/zip' }))).toBe(true)
  })
})

describe('parseStrEmpty', () => {
  it('undefined/null 字符串转空串', () => {
    expect(parseStrEmpty('undefined')).toBe('')
    expect(parseStrEmpty('null')).toBe('')
    expect(parseStrEmpty('abc')).toBe('abc')
  })
})
