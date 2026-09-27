// 通用工具函数 —— 对位 RuoYi-Vue3 src/utils/ruoyi.js，逐行对齐行为
// 注：resetForm 依赖 Vue 实例 $refs，React 版无对应物，不移植（页面内用 antd Form resetFields）。

/**
 * 日期格式化
 * @param time Date | 时间戳(秒/毫秒) | 日期字符串
 * @param pattern 格式串，默认 '{y}-{m}-{d} {h}:{i}:{s}'；{a} 为星期（中文）
 */
export function parseTime(time?: unknown, pattern?: string): string | null {
  if (arguments.length === 0 || !time) {
    return null
  }
  const format = pattern || '{y}-{m}-{d} {h}:{i}:{s}'
  let date: Date
  if (typeof time === 'object') {
    date = time as Date
  } else {
    if (typeof time === 'string' && /^[0-9]+$/.test(time)) {
      time = parseInt(time)
    } else if (typeof time === 'string') {
      // iOS/Safari 兼容：2024-01-01T00:00:00.000 → 2024/01/01 00:00:00
      time = time
        .replace(new RegExp(/-/gm), '/')
        .replace('T', ' ')
        .replace(new RegExp(/\.[\d]{3}/gm), '')
    }
    if (typeof time === 'number' && time.toString().length === 10) {
      time = time * 1000
    }
    date = new Date(time as string | number)
  }
  const formatObj = {
    y: date.getFullYear(),
    m: date.getMonth() + 1,
    d: date.getDate(),
    h: date.getHours(),
    i: date.getMinutes(),
    s: date.getSeconds(),
    a: date.getDay(),
  }
  const time_str = format.replace(/{(y|m|d|h|i|s|a)+}/g, (result, key: string): string => {
    const k = key as 'y' | 'm' | 'd' | 'h' | 'i' | 's' | 'a'
    let value: string | number = formatObj[k]
    // getDay() 周日返回 0
    if (k === 'a') {
      return ['日', '一', '二', '三', '四', '五', '六'][value]
    }
    if (result.length > 0 && value < 10) {
      value = '0' + value
    }
    return String(value || 0)
  })
  return time_str
}

/**
 * 添加日期范围到查询参数（params.beginTime/endTime 或 params.beginXxx/endXxx）
 */
export function addDateRange(
  params: Record<string, unknown>,
  dateRange: unknown[] | undefined,
  propName?: string,
): Record<string, unknown> {
  const search = params
  search.params =
    typeof search.params === 'object' && search.params !== null && !Array.isArray(search.params)
      ? search.params
      : {}
  const range = Array.isArray(dateRange) ? dateRange : []
  const target = search.params as Record<string, unknown>
  if (typeof propName === 'undefined') {
    target['beginTime'] = range[0]
    target['endTime'] = range[1]
  } else {
    target['begin' + propName] = range[0]
    target['end' + propName] = range[1]
  }
  return search
}

/**
 * 回显数据字典 label
 */
export function selectDictLabel<T extends { label: string; value: unknown }>(
  datas: T[],
  value: unknown,
): string {
  if (value === undefined) {
    return ''
  }
  const actions: string[] = []
  datas.some((item) => {
    if (item.value == ('' + value)) {
      actions.push(item.label)
      return true
    }
  })
  if (actions.length === 0) {
    actions.push(value as string)
  }
  return actions.join('')
}

/**
 * 回显数据字典 label（多值，separator 分隔）
 */
export function selectDictLabels<T extends { label: string; value: unknown }>(
  datas: T[],
  value: unknown,
  separator?: string,
): string {
  if (value === undefined || (value as { length: number }).length === 0) {
    return ''
  }
  if (Array.isArray(value)) {
    value = value.join(',')
  }
  const actions: string[] = []
  const currentSeparator = separator === undefined ? ',' : separator
  const temp = String(value).split(currentSeparator)
  temp.some((val) => {
    let match = false
    datas.some((item) => {
      if (item.value == ('' + val)) {
        actions.push(item.label + currentSeparator)
        match = true
      }
    })
    if (!match) {
      actions.push(val + currentSeparator)
    }
  })
  return actions.join('').substring(0, actions.join('').length - 1)
}

/**
 * 转换字符串，undefined/null 等转化为 ""（user.js getUser 的尾斜杠拼接依赖它）
 */
export function parseStrEmpty(str: string): string {
  if (!str || str === 'undefined' || str === 'null') {
    return ''
  }
  return str
}

/**
 * 构造树型结构数据
 * @param data 数据源（会被原地写入 children 字段，与基准一致）
 * @param id id 字段名，默认 'id'
 * @param parentId 父节点字段名，默认 'parentId'
 * @param children 孩子字段名，默认 'children'
 */
export function handleTree<T extends Record<string, any>>(
  data: T[],
  id?: string,
  parentId?: string,
  children?: string,
): T[] {
  const config = {
    id: id || 'id',
    parentId: parentId || 'parentId',
    childrenList: children || 'children',
  }

  const childrenListMap: Record<string, T> = {}
  const tree: T[] = []
  for (const d of data) {
    const key = String(d[config.id])
    childrenListMap[key] = d
    if (!d[config.childrenList]) {
      ;(d as Record<string, unknown>)[config.childrenList] = []
    }
  }

  for (const d of data) {
    const pid = String(d[config.parentId])
    const parentObj = childrenListMap[pid]
    if (!parentObj) {
      tree.push(d)
    } else {
      ;((parentObj as Record<string, unknown>)[config.childrenList] as unknown[]).push(d)
    }
  }
  return tree
}

/**
 * 参数序列化（GET 拼接 / 下载表单编码共用）
 * 嵌套对象展开为 propName[key]=value；跳过 null/''/undefined；每段以 & 结尾（调用方按需去掉尾 &，与基准一致）
 */
export function tansParams(params: Record<string, unknown>): string {
  let result = ''
  for (const propName of Object.keys(params)) {
    const value = params[propName]
    const part = encodeURIComponent(propName) + '='
    if (value !== null && value !== '' && typeof value !== 'undefined') {
      if (typeof value === 'object') {
        for (const key of Object.keys(value as Record<string, unknown>)) {
          const sub = (value as Record<string, unknown>)[key]
          if (sub !== null && sub !== '' && typeof sub !== 'undefined') {
            const param = propName + '[' + key + ']'
            const subPart = encodeURIComponent(param) + '='
            result += subPart + encodeURIComponent(String(sub)) + '&'
          }
        }
      } else {
        result += part + encodeURIComponent(String(value)) + '&'
      }
    }
  }
  return result
}

/**
 * 归一化路径（双斜杠合并、去尾斜杠）
 */
export function getNormalPath(p: string): string {
  if (p.length === 0 || !p || p === 'undefined') {
    return p
  }
  let res = p.replace('//', '/')
  if (res[res.length - 1] === '/') {
    return res.slice(0, res.length - 1)
  }
  return res
}

/**
 * 验证是否为 blob 文件（否则是 JSON 错误体）
 */
export function blobValidate(data: Blob): boolean {
  return data.type !== 'application/json'
}
