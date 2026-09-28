// DictTag 导出的 elTagType → antd color 映射纯数据单测
// (组件本体含 antd Tag/Typography 渲染,devDependencies 无 testing-library,不做渲染测试)
import { describe, expect, it } from 'vitest'
import { TAG_TYPE_COLOR } from './index'

describe('TAG_TYPE_COLOR(基准 elTagType → antd color 映射)', () => {
  it('五类语义色与基准一一对应', () => {
    expect(TAG_TYPE_COLOR['']).toBe('default')
    expect(TAG_TYPE_COLOR.default).toBe('default')
    expect(TAG_TYPE_COLOR.success).toBe('green')
    expect(TAG_TYPE_COLOR.info).toBe('blue')
    expect(TAG_TYPE_COLOR.warning).toBe('orange')
    expect(TAG_TYPE_COLOR.danger).toBe('red')
  })

  it('primary 归一为 blue(基准 default 才是主色,antd 主色用 blue 表达)', () => {
    expect(TAG_TYPE_COLOR.primary).toBe('blue')
  })

  it('键集合封闭:恰好 7 个(防误增漏改)', () => {
    expect(Object.keys(TAG_TYPE_COLOR).sort()).toEqual(
      ['', 'danger', 'default', 'info', 'primary', 'success', 'warning'].sort(),
    )
  })

  it('消费端语义:未登记类型回退 default(TAG_TYPE_COLOR[hit.elTagType || "default"] || "default" 的兜底链)', () => {
    // DictTag 组件内取色逻辑:TAG_TYPE_COLOR[type] 未命中 → 'default'
    const colorOf = (elTagType?: string) => TAG_TYPE_COLOR[elTagType || 'default'] || 'default'
    expect(colorOf(undefined)).toBe('default')
    expect(colorOf('not-a-type')).toBe('default')
    expect(colorOf('danger')).toBe('red')
  })
})
