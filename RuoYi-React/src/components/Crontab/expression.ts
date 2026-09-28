// Crontab 表达式纯函数层 —— 对位基准 RuoYi-Vue3/src/components/Crontab 各域组件的拼接/反解析逻辑
// 基准中这些逻辑分散在 second/min/hour/day/month/week/year.vue 的 computed + watch 里，
// 此处集中为纯函数便于单测；组件层（FieldPanel.tsx）只做交互，生成片段必须与基准逐字等价。

/** 七域表达式值（对位基准 index.vue crontabValueObj） */
export interface CronValue {
  second: string
  min: string
  hour: string
  day: string
  month: string
  week: string
  year: string
}

/** 默认表达式（对位基准 crontabValueObj 初值 / clearCron） */
export const DEFAULT_CRON: CronValue = {
  second: '*',
  min: '*',
  hour: '*',
  day: '*',
  month: '*',
  week: '?',
  year: '',
}

/** 数字钳位（对位基准 index.vue checkNumber：向下取整后夹逼） */
export function checkNumber(value: number, minLimit: number, maxLimit: number): number {
  const floored = Math.floor(value)
  if (floored < minLimit) return minLimit
  if (floored > maxLimit) return maxLimit
  return floored
}

/** 七段 + 年域拼接（对位基准 crontabValueString：年域为空串时不追加） */
export function buildCronExpression(v: CronValue): string {
  return (
    v.second +
    ' ' +
    v.min +
    ' ' +
    v.hour +
    ' ' +
    v.day +
    ' ' +
    v.month +
    ' ' +
    v.week +
    (v.year === '' ? '' : ' ' + v.year)
  )
}

/** 表达式反解析（对位基准 resolveExp：按空白切分，>=6 段才合法，年域可缺省；不合法返回 null 保持现状） */
export function parseCronExpression(expression: string): CronValue | null {
  const arr = expression.split(/\s+/)
  if (arr.length >= 6) {
    return {
      second: arr[0],
      min: arr[1],
      hour: arr[2],
      day: arr[3],
      month: arr[4],
      week: arr[5],
      year: arr[6] ?? '',
    }
  }
  return null
}

/** 指定列表反解析（对位基准 changeRadioValue else 分支：Number 化 + Set 去重） */
function parseAssignList(value: string): number[] {
  return [...new Set(value.split(',').map((item) => Number(item)))]
}

// ============================== 秒 / 分 / 时 / 月（同构四域） ==============================

/** 同构四域的 UI 状态（对位各 vue 组件的 radioValue + cycle01/02 + average01/02 + checkboxList） */
export interface SimpleFieldState {
  /** 1 通配 | 2 周期 | 3 从X开始每Y | 4 指定 */
  radio: number
  cycle01: number
  cycle02: number
  average01: number
  average02: number
  checkboxList: number[]
}

/** 同构四域取值范围（base：下界，秒/分/时为 0、月为 1；unitMax：该域最大值，如秒 59、时 23、月 12） */
export interface SimpleFieldRange {
  base: number
  unitMax: number
}

export const SECOND_MIN_RANGE: SimpleFieldRange = { base: 0, unitMax: 59 }
export const HOUR_RANGE: SimpleFieldRange = { base: 0, unitMax: 23 }
export const MONTH_RANGE: SimpleFieldRange = { base: 1, unitMax: 12 }

/** 同构四域初始状态（对位各 vue 组件 ref 初值） */
export function simpleFieldInit(range: SimpleFieldRange): SimpleFieldState {
  return {
    radio: 1,
    cycle01: range.base,
    cycle02: range.base + 1,
    average01: range.base,
    average02: 1,
    checkboxList: [],
  }
}

/** 片段 → 状态（对位 changeRadioValue：只覆写命中分支的字段，其余保持 prev——与 Vue ref 持久语义一致） */
export function parseSimpleSegment(
  value: string,
  _range: SimpleFieldRange,
  prev: SimpleFieldState,
): SimpleFieldState {
  if (value === '*') return { ...prev, radio: 1 }
  if (value.indexOf('-') > -1) {
    const arr = value.split('-')
    return { ...prev, radio: 2, cycle01: Number(arr[0]), cycle02: Number(arr[1]) }
  }
  if (value.indexOf('/') > -1) {
    const arr = value.split('/')
    return { ...prev, radio: 3, average01: Number(arr[0]), average02: Number(arr[1]) }
  }
  return { ...prev, radio: 4, checkboxList: parseAssignList(value) }
}

/** 状态 → 片段（对位 onRadioChange + cycleTotal/averageTotal/checkboxString 的钳位拼接） */
export function buildSimpleSegment(st: SimpleFieldState, range: SimpleFieldRange): string {
  const { base, unitMax } = range
  switch (st.radio) {
    case 2: {
      const c01 = checkNumber(st.cycle01, base, unitMax - 1)
      const c02 = checkNumber(st.cycle02, c01 + 1, unitMax)
      return c01 + '-' + c02
    }
    case 3: {
      const a01 = checkNumber(st.average01, base, unitMax - 1)
      const a02 = checkNumber(st.average02, 1, unitMax - a01)
      return a01 + '/' + a02
    }
    case 4:
      return st.checkboxList.join(',')
    case 1:
    default:
      return '*'
  }
}

// ============================================== 日 ==============================================

/** 日域 UI 状态（对位 day.vue） */
export interface DayFieldState {
  /** 1 * | 2 ? 不指定 | 3 周期 | 4 从X开始每Y | 5 工作日XW | 6 本月最后 L | 7 指定 */
  radio: number
  cycle01: number
  cycle02: number
  average01: number
  average02: number
  workday: number
  checkboxList: number[]
}

export function dayFieldInit(): DayFieldState {
  return { radio: 1, cycle01: 1, cycle02: 2, average01: 1, average02: 1, workday: 1, checkboxList: [] }
}

/** 日域片段 → 状态（对位 day.vue changeRadioValue，分支顺序逐字对位） */
export function parseDaySegment(value: string, prev: DayFieldState): DayFieldState {
  if (value === '*') return { ...prev, radio: 1 }
  if (value === '?') return { ...prev, radio: 2 }
  if (value.indexOf('-') > -1) {
    const arr = value.split('-')
    return { ...prev, radio: 3, cycle01: Number(arr[0]), cycle02: Number(arr[1]) }
  }
  if (value.indexOf('/') > -1) {
    const arr = value.split('/')
    return { ...prev, radio: 4, average01: Number(arr[0]), average02: Number(arr[1]) }
  }
  if (value.indexOf('W') > -1) {
    return { ...prev, radio: 5, workday: Number(value.split('W')[0]) }
  }
  if (value === 'L') return { ...prev, radio: 6 }
  return { ...prev, radio: 7, checkboxList: parseAssignList(value) }
}

/** 日域状态 → 片段（对位 day.vue onRadioChange + 各 computed 钳位拼接） */
export function buildDaySegment(st: DayFieldState): string {
  switch (st.radio) {
    case 2:
      return '?'
    case 3: {
      const c01 = checkNumber(st.cycle01, 1, 30)
      const c02 = checkNumber(st.cycle02, c01 + 1, 31)
      return c01 + '-' + c02
    }
    case 4: {
      const a01 = checkNumber(st.average01, 1, 30)
      const a02 = checkNumber(st.average02, 1, 31 - a01)
      return a01 + '/' + a02
    }
    case 5:
      return checkNumber(st.workday, 1, 31) + 'W'
    case 6:
      return 'L'
    case 7:
      return st.checkboxList.join(',')
    case 1:
    default:
      return '*'
  }
}

// ============================================== 周 ==============================================

/** 周域 UI 状态（对位 week.vue；average01=第 N 周、average02=星期 X） */
export interface WeekFieldState {
  /** 1 * | 2 ? 不指定 | 3 周期 | 4 第N周的星期X | 5 本月最后一个星期X | 6 指定 */
  radio: number
  cycle01: number
  cycle02: number
  average01: number
  average02: number
  weekday: number
  checkboxList: number[]
}

/** 基准 week.vue 初值：radio 默认 2（不指定）、cycle01=2、cycle02=3 */
export function weekFieldInit(): WeekFieldState {
  return { radio: 2, cycle01: 2, cycle02: 3, average01: 1, average02: 2, weekday: 2, checkboxList: [] }
}

/** 周域片段 → 状态（对位 week.vue changeRadioValue；'#' 拆分时 average01=井号后、average02=井号前） */
export function parseWeekSegment(value: string, prev: WeekFieldState): WeekFieldState {
  if (value === '*') return { ...prev, radio: 1 }
  if (value === '?') return { ...prev, radio: 2 }
  if (value.indexOf('-') > -1) {
    const arr = value.split('-')
    return { ...prev, radio: 3, cycle01: Number(arr[0]), cycle02: Number(arr[1]) }
  }
  if (value.indexOf('#') > -1) {
    const arr = value.split('#')
    return { ...prev, radio: 4, average01: Number(arr[1]), average02: Number(arr[0]) }
  }
  if (value.indexOf('L') > -1) {
    return { ...prev, radio: 5, weekday: Number(value.split('L')[0]) }
  }
  return { ...prev, radio: 6, checkboxList: parseAssignList(value) }
}

/** 周域状态 → 片段（对位 week.vue onRadioChange；第N周拼为「星期X#第N周」） */
export function buildWeekSegment(st: WeekFieldState): string {
  switch (st.radio) {
    case 3: {
      const c01 = checkNumber(st.cycle01, 1, 6)
      const c02 = checkNumber(st.cycle02, c01 + 1, 7)
      return c01 + '-' + c02
    }
    case 4: {
      const a01 = checkNumber(st.average01, 1, 4)
      const a02 = checkNumber(st.average02, 1, 7)
      return a02 + '#' + a01
    }
    case 5:
      return checkNumber(st.weekday, 1, 7) + 'L'
    case 6:
      return st.checkboxList.join(',')
    case 2:
      return '?'
    case 1:
    default:
      return '*'
  }
}

// ============================================== 年 ==============================================

/** 年域 UI 状态（对位 year.vue；fullYear 为组件实例化当年） */
export interface YearFieldState {
  /** 1 不填 | 2 每年 | 3 周期 | 4 从X开始每Y | 5 指定 */
  radio: number
  cycle01: number
  cycle02: number
  average01: number
  average02: number
  checkboxList: number[]
}

export function yearFieldInit(fullYear: number): YearFieldState {
  return {
    radio: 1,
    cycle01: fullYear,
    cycle02: fullYear + 1,
    average01: fullYear,
    average02: 1,
    checkboxList: [],
  }
}

/** 年域片段 → 状态（对位 year.vue changeRadioValue；'' 为不填） */
export function parseYearSegment(value: string, prev: YearFieldState): YearFieldState {
  if (value === '') return { ...prev, radio: 1 }
  if (value === '*') return { ...prev, radio: 2 }
  if (value.indexOf('-') > -1) {
    const arr = value.split('-')
    return { ...prev, radio: 3, cycle01: Number(arr[0]), cycle02: Number(arr[1]) }
  }
  if (value.indexOf('/') > -1) {
    const arr = value.split('/')
    return { ...prev, radio: 4, average01: Number(arr[0]), average02: Number(arr[1]) }
  }
  return { ...prev, radio: 5, checkboxList: parseAssignList(value) }
}

/** 年域状态 → 片段（对位 year.vue onRadioChange；范围 [fullYear, fullYear+10]，步长上限 10） */
export function buildYearSegment(st: YearFieldState, fullYear: number): string {
  const maxFullYear = fullYear + 10
  switch (st.radio) {
    case 2:
      return '*'
    case 3: {
      const c01 = checkNumber(st.cycle01, fullYear, maxFullYear - 1)
      const c02 = checkNumber(st.cycle02, c01 + 1, maxFullYear)
      return c01 + '-' + c02
    }
    case 4: {
      const a01 = checkNumber(st.average01, fullYear, maxFullYear - 1)
      const a02 = checkNumber(st.average02, 1, 10)
      return a01 + '/' + a02
    }
    case 5:
      return st.checkboxList.join(',')
    case 1:
    default:
      return ''
  }
}

// ====================================== 日 / 周互斥联动 ======================================

/**
 * 日 → 周互斥联动（对位 day.vue onRadioChange 前置逻辑）：
 * 日选「不指定(?)」时周若也是 ? 则周改 *；日选其他时周若非 ? 则周改 ?。返回 null 表示无需联动。
 */
export function weekAdjustFromDay(dayRadio: number, week: string): string | null {
  if (dayRadio === 2 && week === '?') return '*'
  if (dayRadio !== 2 && week !== '?') return '?'
  return null
}

/** 周 → 日互斥联动（对位 week.vue onRadioChange 前置逻辑），语义与上对称 */
export function dayAdjustFromWeek(weekRadio: number, day: string): string | null {
  if (weekRadio === 2 && day === '?') return '*'
  if (weekRadio !== 2 && day !== '?') return '?'
  return null
}
