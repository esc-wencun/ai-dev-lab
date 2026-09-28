// Crontab expression 纯函数单测 —— 覆盖各域主要分支 + 日周互斥 + 完整表达式拼接
// 断言对照基准 RuoYi-Vue3/src/components/Crontab 各域 vue 的 computed 拼接与 changeRadioValue 反解析行为

import { describe, expect, it } from 'vitest'
import {
  DEFAULT_CRON,
  buildCronExpression,
  buildDaySegment,
  buildSimpleSegment,
  buildWeekSegment,
  buildYearSegment,
  checkNumber,
  dayAdjustFromWeek,
  parseCronExpression,
  parseDaySegment,
  parseSimpleSegment,
  parseWeekSegment,
  parseYearSegment,
  weekAdjustFromDay,
  type CronValue,
  type DayFieldState,
  type SimpleFieldState,
  type WeekFieldState,
  type YearFieldState,
} from './expression'

const SEC = { base: 0, unitMax: 59 }
const HOUR = { base: 0, unitMax: 23 }
const MONTH = { base: 1, unitMax: 12 }

function simpleInit(base: number): SimpleFieldState {
  return { radio: 1, cycle01: base, cycle02: base + 1, average01: base, average02: 1, checkboxList: [] }
}
function dayInit(): DayFieldState {
  return { radio: 1, cycle01: 1, cycle02: 2, average01: 1, average02: 1, workday: 1, checkboxList: [] }
}
function weekInit(): WeekFieldState {
  return { radio: 2, cycle01: 2, cycle02: 3, average01: 1, average02: 2, weekday: 2, checkboxList: [] }
}
function yearInit(fullYear: number): YearFieldState {
  return { radio: 1, cycle01: fullYear, cycle02: fullYear + 1, average01: fullYear, average02: 1, checkboxList: [] }
}

describe('checkNumber（对位基准 index.vue checkNumber）', () => {
  it('范围内整数原样返回', () => {
    expect(checkNumber(30, 0, 59)).toBe(30)
  })
  it('向下取整', () => {
    expect(checkNumber(12.9, 0, 59)).toBe(12)
  })
  it('越界钳位', () => {
    expect(checkNumber(-5, 1, 30)).toBe(1)
    expect(checkNumber(99, 1, 30)).toBe(30)
  })
})

describe('buildCronExpression（七段 + 年域拼接）', () => {
  it('年域为空串时不追加第七段', () => {
    expect(buildCronExpression(DEFAULT_CRON)).toBe('* * * * * ?')
  })
  it('年域非空时追加第七段', () => {
    const v: CronValue = { ...DEFAULT_CRON, year: '2026' }
    expect(buildCronExpression(v)).toBe('* * * * * ? 2026')
  })
})

describe('parseCronExpression（对位基准 resolveExp）', () => {
  it('6 段表达式正常解析，年域缺省为空串', () => {
    expect(parseCronExpression('0 0 8 * * ?')).toEqual({
      second: '0', min: '0', hour: '8', day: '*', month: '*', week: '?', year: '',
    })
  })
  it('7 段表达式解析年域', () => {
    expect(parseCronExpression('0/5 1-5 1/2 1L 1-3 2#1 2026-2030')?.year).toBe('2026-2030')
  })
  it('不足 6 段返回 null（维持现状，对位基准不覆盖 crontabValueObj）', () => {
    expect(parseCronExpression('0 0 8')).toBeNull()
    expect(parseCronExpression('')).toBeNull()
  })
})

describe('秒/分/时/月同构域', () => {
  it('parse：* → 通配', () => {
    expect(parseSimpleSegment('*', SEC, simpleInit(0)).radio).toBe(1)
  })
  it('parse：周期片段', () => {
    const st = parseSimpleSegment('10-30', SEC, simpleInit(0))
    expect(st.radio).toBe(2)
    expect(st.cycle01).toBe(10)
    expect(st.cycle02).toBe(30)
  })
  it('parse：从X开始每Y', () => {
    const st = parseSimpleSegment('5/10', HOUR, simpleInit(0))
    expect(st.radio).toBe(3)
    expect(st.average01).toBe(5)
    expect(st.average02).toBe(10)
  })
  it('parse：指定列表 Number 化 + 去重', () => {
    const st = parseSimpleSegment('1,1,3', SEC, simpleInit(0))
    expect(st.radio).toBe(4)
    expect(st.checkboxList).toEqual([1, 3])
  })
  it('build：通配 / 周期 / 步长 / 指定（含钳位）', () => {
    expect(buildSimpleSegment(simpleInit(0), SEC)).toBe('*')
    // cycle01=70 钳到 58（上界 unitMax-1），cycle02=3 随之钳到 [59,59]
    expect(buildSimpleSegment({ ...simpleInit(0), radio: 2, cycle01: 70, cycle02: 3 }, SEC)).toBe('58-59')
    expect(buildSimpleSegment({ ...simpleInit(0), radio: 3, average01: 0, average02: 59 }, SEC)).toBe('0/59')
    expect(
      buildSimpleSegment({ ...simpleInit(1), radio: 4, checkboxList: [3, 1, 2] }, MONTH),
    ).toBe('3,1,2') // join 保序，不排序（对位基准 checkboxList.join(',')）
  })
  it('build：周期起点钳位后联动终点（cycle01=58 → cycle02 ∈ [59,59]）', () => {
    const seg = buildSimpleSegment({ ...simpleInit(0), radio: 2, cycle01: 58, cycle02: 5 }, SEC)
    expect(seg).toBe('58-59')
  })
  it('build：步长上限随 average01 收缩（hour：average01=23 越界钳到 22 → a02=1 → 22/1）', () => {
    const seg = buildSimpleSegment({ ...simpleInit(0), radio: 3, average01: 23, average02: 5 }, HOUR)
    expect(seg).toBe('22/1')
  })
})

describe('日域', () => {
  it('parse：7 种片段各归位', () => {
    expect(parseDaySegment('*', dayInit()).radio).toBe(1)
    expect(parseDaySegment('?', dayInit()).radio).toBe(2)
    const cyc = parseDaySegment('5-20', dayInit())
    expect(cyc.radio).toBe(3)
    expect(cyc.cycle01).toBe(5)
    const avg = parseDaySegment('10/5', dayInit())
    expect(avg.radio).toBe(4)
    expect(avg.average02).toBe(5)
    const wd = parseDaySegment('15W', dayInit())
    expect(wd.radio).toBe(5)
    expect(wd.workday).toBe(15)
    expect(parseDaySegment('L', dayInit()).radio).toBe(6)
    expect(parseDaySegment('1,15', dayInit()).radio).toBe(7)
  })
  it('build：工作日 XW / 本月最后 L / 周期', () => {
    expect(buildDaySegment({ ...dayInit(), radio: 5, workday: 15 })).toBe('15W')
    expect(buildDaySegment({ ...dayInit(), radio: 6 })).toBe('L')
    expect(buildDaySegment({ ...dayInit(), radio: 3, cycle01: 1, cycle02: 31 })).toBe('1-31')
    expect(buildDaySegment({ ...dayInit(), radio: 2 })).toBe('?')
    expect(buildDaySegment(dayInit())).toBe('*')
  })
})

describe('周域', () => {
  it('parse：# 片段 average01=井号后(第N周)、average02=井号前(星期X)', () => {
    const st = parseWeekSegment('2#3', weekInit())
    expect(st.radio).toBe(4)
    expect(st.average01).toBe(3)
    expect(st.average02).toBe(2)
  })
  it('parse：XL 片段 / 周期 / 指定', () => {
    const last = parseWeekSegment('6L', weekInit())
    expect(last.radio).toBe(5)
    expect(last.weekday).toBe(6)
    const cyc = parseWeekSegment('2-5', weekInit())
    expect(cyc.radio).toBe(3)
    expect(cyc.cycle02).toBe(5)
    expect(parseWeekSegment('1,7', weekInit()).radio).toBe(6)
  })
  it('build：第N周拼为 星期X#N', () => {
    expect(buildWeekSegment({ ...weekInit(), radio: 4, average01: 3, average02: 2 })).toBe('2#3')
  })
  it('build：周期 / 本月最后星期X / 不指定 / 通配', () => {
    expect(buildWeekSegment({ ...weekInit(), radio: 3, cycle01: 2, cycle02: 3 })).toBe('2-3')
    expect(buildWeekSegment({ ...weekInit(), radio: 5, weekday: 6 })).toBe('6L')
    expect(buildWeekSegment(weekInit())).toBe('?')
    expect(buildWeekSegment({ ...weekInit(), radio: 1 })).toBe('*')
  })
})

describe('年域', () => {
  it('parse：空串/每年/周期/步长/指定', () => {
    expect(parseYearSegment('', yearInit(2026)).radio).toBe(1)
    expect(parseYearSegment('*', yearInit(2026)).radio).toBe(2)
    const cyc = parseYearSegment('2026-2030', yearInit(2026))
    expect(cyc.radio).toBe(3)
    expect(cyc.cycle01).toBe(2026)
    const avg = parseYearSegment('2026/2', yearInit(2026))
    expect(avg.radio).toBe(4)
    expect(parseYearSegment('2026,2028', yearInit(2026)).radio).toBe(5)
  })
  it('build：不填为空串、每年为 *、周期钳位在 [当年, 当年+10]', () => {
    expect(buildYearSegment(yearInit(2026), 2026)).toBe('')
    expect(buildYearSegment({ ...yearInit(2026), radio: 2 }, 2026)).toBe('*')
    expect(buildYearSegment({ ...yearInit(2026), radio: 3, cycle01: 2026, cycle02: 2036 }, 2026)).toBe('2026-2036')
    expect(buildYearSegment({ ...yearInit(2026), radio: 4, average01: 2026, average02: 2 }, 2026)).toBe('2026/2')
    expect(buildYearSegment({ ...yearInit(2026), radio: 5, checkboxList: [2026, 2028] }, 2026)).toBe('2026,2028')
  })
})

describe('日 / 周互斥联动（对位基准 day.vue / week.vue onRadioChange 前置逻辑）', () => {
  it('day 选 ?（不指定）且 week 为 ? → week 变 *', () => {
    expect(weekAdjustFromDay(2, '?')).toBe('*')
  })
  it('day 选 ? 且 week 非 ?（如 *）→ week 保持不动', () => {
    expect(weekAdjustFromDay(2, '*')).toBeNull()
  })
  it('day 选其他且 week 非 ? → week 变 ?', () => {
    expect(weekAdjustFromDay(1, '2-4')).toBe('?')
  })
  it('day 选其他且 week 为 ? → week 保持不动', () => {
    expect(weekAdjustFromDay(1, '?')).toBeNull()
  })
  it('week 侧对称：week 选 ? 且 day 为 ? → day 变 *；week 选其他且 day 非 ? → day 变 ?', () => {
    expect(dayAdjustFromWeek(2, '?')).toBe('*')
    expect(dayAdjustFromWeek(2, '15W')).toBeNull()
    expect(dayAdjustFromWeek(4, '*')).toBe('?')
    expect(dayAdjustFromWeek(4, '?')).toBeNull()
  })
})
