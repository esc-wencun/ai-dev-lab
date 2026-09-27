// @vitest-environment jsdom
import { describe, expect, it } from 'vitest'

// 密码规则单测 —— 断言对照 RuoYi-Vue3 src/utils/passwordRule.js
// 模块级读取 sessionStorage，故用 jsdom 环境注释指令

describe('passwordRule', () => {
  it('规则表与基准一致（0-4 五条）', async () => {
    sessionStorage.clear()
    const mod = await import('./passwordRule')
    const r0 = mod.getRule('0')
    expect(r0.pattern.test('abc123')).toBe(true)
    expect(r0.pattern.test('a<b')).toBe(false)
    expect(r0.message).toBe('密码不能包含非法字符：< > " \' \\ |')

    const r1 = mod.getRule('1')
    expect(r1.pattern.test('123456')).toBe(true)
    expect(r1.pattern.test('123456a')).toBe(false)

    const r2 = mod.getRule('2')
    expect(r2.pattern.test('abcABC')).toBe(true)
    expect(r2.pattern.test('abc123')).toBe(false)

    const r3 = mod.getRule('3')
    expect(r3.pattern.test('abc123')).toBe(true)
    expect(r3.pattern.test('abcdef')).toBe(false)
    expect(r3.pattern.test('123456')).toBe(false)

    const r4 = mod.getRule('4')
    expect(r4.pattern.test('abc123!')).toBe(true) // ! 在特殊字符集内
    expect(r4.pattern.test('abc123_')).toBe(true)
    expect(r4.pattern.test('abc123')).toBe(false) // 缺特殊字符
    expect(r4.pattern.test('abc12!')).toBe(true) // 有数字有字母有特殊字符
    expect(r4.pattern.test('abcdef!')).toBe(false) // 缺数字
  })

  it('非法字符集 < > " \' \\ | 被规则0拒绝', async () => {
    sessionStorage.clear()
    const mod = await import('./passwordRule')
    const r0 = mod.getRule('0')
    for (const ch of ['<', '>', '"', "'", '\\', '|']) {
      expect(r0.pattern.test('a' + ch)).toBe(false)
    }
  })

  it('pwdValidatorFactory：长度与非法字符校验', async () => {
    sessionStorage.clear()
    const mod = await import('./passwordRule')
    const v = mod.pwdValidatorFactory('login')
    await expect(v(null, 'a<b123')).rejects.toThrow('密码不能包含非法字符：< > " \' \\ |')
    await expect(v(null, '1')).rejects.toThrow('密码长度必须介于 6 和 20 之间')
    await expect(v(null, '')).rejects.toThrow('密码不能为空')
    await expect(v(null, 'abc123')).resolves.toBeUndefined()
  })

  it('pwdValidatorFactory register：固定规则0 + 前缀「用户密码」', async () => {
    sessionStorage.clear()
    const mod = await import('./passwordRule')
    const v = mod.pwdValidatorFactory('register')
    await expect(v(null, '123456')).resolves.toBeUndefined()
    await expect(v(null, 'abc123')).resolves.toBeUndefined()
    await expect(v(null, '1')).rejects.toThrow('用户密码长度必须介于 6 和 20 之间')
  })
})
