// 密码强度规则 —— 对位 RuoYi-Vue3 src/utils/passwordRule.js
// 根据 sessionStorage 键 `pwrChrtype`（基准键名即如此拼写，非笔误）动态选择规则

import cache from '@/plugins/cache'

// 密码限制类型（登录后 getInfo 写入，默认 '0'）
const pwdChrType = cache.session.get('pwrChrtype') || '0'

// 各类型对应的正则与错误提示（与基准逐字一致）
const PWD_RULES: Record<string, { pattern: RegExp; message: string }> = {
  '0': { pattern: /^[^<>"'|\\]+$/, message: '密码不能包含非法字符：< > " \' \\ |' },
  '1': { pattern: /^[0-9]+$/, message: '密码只能为数字（0-9）' },
  '2': { pattern: /^[a-zA-Z]+$/, message: '密码只能为英文字母（a-z、A-Z）' },
  '3': { pattern: /^(?=.*[a-zA-Z])(?=.*[0-9])[a-zA-Z0-9]+$/, message: '密码必须同时包含字母和数字' },
  '4': {
    pattern: /^(?=.*[A-Za-z])(?=.*\d)(?=.*[~!@#$%^&*()\-=_+])[A-Za-z\d~!@#$%^&*()\-=_+]+$/,
    message: '密码必须同时包含字母、数字和特殊字符（~!@#$%^&*()-=_+）',
  },
}

export function getPwdChrType(): string {
  return pwdChrType
}

export function getRule(type?: string): { pattern: RegExp; message: string } {
  return PWD_RULES[type || pwdChrType] || PWD_RULES['0']
}

// 表单校验工厂（供 antd Form rule.validator 使用）
// kind: 'login' 通用 | 'info' 个人中心新密码（前缀「新密码」）| 'register' 注册（固定规则'0'，前缀「用户密码」）
export function pwdValidatorFactory(kind: 'login' | 'info' | 'register' = 'login') {
  const rule = getRule()
  const prefix = kind === 'info' ? '新密码' : kind === 'register' ? '用户密码' : '密码'
  return (_rule: unknown, value: string): Promise<void> => {
    if (!value) {
      return Promise.reject(new Error(kind === 'register' ? '请输入您的密码' : `${prefix}不能为空`))
    }
    if (value.length < 6 || value.length > 20) {
      return Promise.reject(new Error(`${prefix}长度必须介于 6 和 20 之间`))
    }
    const r = kind === 'register' ? PWD_RULES['0'] : rule
    if (!r.pattern.test(value)) {
      return Promise.reject(new Error(r.message))
    }
    return Promise.resolve()
  }
}
