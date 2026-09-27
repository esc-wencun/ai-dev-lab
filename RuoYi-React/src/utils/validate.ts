// 校验工具 —— 对位 RuoYi-Vue3 src/utils/validate.js（只移植本项目消费的函数，行为逐行对齐）

/**
 * 路径匹配器（守卫白名单用）
 * @param pattern 支持 `*`→[^/]*、`**`→.*、`?`→[^/]，其余正则字符转义
 */
export function isPathMatch(pattern: string, path: string): boolean {
  const regexPattern = pattern
    .replace(/([.+^${}()|[\]\\])/g, '\\$1')
    .replace(/\*\*/g, '__DOUBLE_STAR__')
    .replace(/\*/g, '[^/]*')
    .replace(/__DOUBLE_STAR__/g, '.*')
    .replace(/\?/g, '[^/]')
  const regex = new RegExp(`^${regexPattern}$`)
  return regex.test(path)
}

/**
 * 判断 value 字符串是否为空
 */
export function isEmpty(value: unknown): boolean {
  if (value == null || value === '' || value === undefined || value === 'undefined') {
    return true
  }
  return false
}

/**
 * 判断 url 是否是 http 或 https
 */
export function isHttp(url: string): boolean {
  return url.indexOf('http://') !== -1 || url.indexOf('https://') !== -1
}

/**
 * 判断 path 是否为外链
 */
export function isExternal(path: string): boolean {
  return /^(https?:|mailto:|tel:)/.test(path)
}
