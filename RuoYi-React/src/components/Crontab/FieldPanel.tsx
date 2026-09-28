// Crontab 七域面板 —— 对位基准 RuoYi-Vue3/src/components/Crontab/{second,min,hour,day,month,week,year}.vue
// 七个域组件高度同构（单选集 + InputNumber/Select 输入），此处合并为面板组件按域分支渲染；
// 文案、选项集合、钳位拼接全部经 expression.ts 纯函数与基准逐字对位（month 的「月月执行一次」、
// hour 周期单位「时」、min「分钟开始， 每」的空格等均保留基准原样）。
// 有意差异：基准靠 watch(cron.xxx) 反解析回显，React 侧由 index.tsx 在挂载时用 parseXxxSegment 初始化状态
// （job 页 destroyOnHidden 每次弹出重挂载，时序等价）；「指定」多选清空时基准会回填上一次选中值，
// antd 多选直接允许清空（生成结果为空串由 job 页校验兜底），见 deviations 说明。

import { InputNumber, Radio, Select } from 'antd'
import {
  checkNumber,
  dayAdjustFromWeek,
  parseDaySegment,
  parseSimpleSegment,
  parseWeekSegment,
  parseYearSegment,
  buildDaySegment,
  buildSimpleSegment,
  buildWeekSegment,
  buildYearSegment,
  weekAdjustFromDay,
  type CronValue,
  type DayFieldState,
  type SimpleFieldRange,
  type SimpleFieldState,
  type WeekFieldState,
  type YearFieldState,
} from './expression'

export type FieldName = 'second' | 'min' | 'hour' | 'day' | 'month' | 'week' | 'year'

/** 周选项（对位基准 week.vue weekList：key 1=星期日 … 7=星期六） */
export const WEEK_LIST = [
  { key: 1, value: '星期日' },
  { key: 2, value: '星期一' },
  { key: 3, value: '星期二' },
  { key: 4, value: '星期三' },
  { key: 5, value: '星期四' },
  { key: 6, value: '星期五' },
  { key: 7, value: '星期六' },
]

/** 月选项（对位基准 month.vue monthList：key 1=一月 … 12=十二月） */
export const MONTH_LIST = [
  { key: 1, value: '一月' },
  { key: 2, value: '二月' },
  { key: 3, value: '三月' },
  { key: 4, value: '四月' },
  { key: 5, value: '五月' },
  { key: 6, value: '六月' },
  { key: 7, value: '七月' },
  { key: 8, value: '八月' },
  { key: 9, value: '九月' },
  { key: 10, value: '十月' },
  { key: 11, value: '十一月' },
  { key: 12, value: '十二月' },
]

/** 指定多选上限（对位基准 multiple-limit：秒/分/时 10、日 10、月/年 8、周 6） */
const ASSIGN_LIMIT: Record<FieldName, number> = {
  second: 10,
  min: 10,
  hour: 10,
  day: 10,
  month: 8,
  week: 6,
  year: 8,
}

interface SimplePanelProps {
  name: 'second' | 'min' | 'hour' | 'month'
  /** 单选 1 文案（对位基准：秒/分钟/小时/月，允许的通配符[, - * /]） */
  radioLabel: string
  /** 周期/从X开始处的单位字（hour 为「时」，其余同 radioLabel 首词） */
  cycleUnit: string
  /** 「每 X」后的文案（对位基准：秒执行一次/分钟执行一次/小时执行一次/月月执行一次） */
  stepLabel: string
  state: SimpleFieldState
  range: SimpleFieldRange
  onSegment: (name: FieldName, value: string) => void
  onState: (name: FieldName, state: SimpleFieldState) => void
}

/** 秒 / 分 / 时 / 月同构面板（对位基准四张同构 vue） */
function SimpleFieldPanel({ name, radioLabel, cycleUnit, stepLabel, state, range, onSegment, onState }: SimplePanelProps) {
  const { base, unitMax } = range
  const set = (patch: Partial<SimpleFieldState>) => {
    const next = { ...state, ...patch }
    onState(name, next)
    onSegment(name, buildSimpleSegment(next, range))
  }
  // 指定下拉选项：月域用中文月名（对位基准 month.vue monthList），其余为数字 0..N
  const assignOptions =
    name === 'month'
      ? MONTH_LIST.map((m) => ({ label: m.value, value: m.key }))
      : seq(base, unitMax).map((n) => ({ label: String(n), value: n }))
  return (
    <div style={{ padding: '12px 24px', display: 'flex', flexDirection: 'column', gap: 12 }}>
      <Radio checked={state.radio === 1} onChange={() => set({ radio: 1 })}>
        {radioLabel}
      </Radio>
      <Radio checked={state.radio === 2} onChange={() => set({ radio: 2 })}>
        周期从
        <NumInput value={state.cycle01} min={base} max={unitMax - 1}
          onChange={(v) => set({ cycle01: v, cycle02: checkNumber(state.cycle02, v + 1, unitMax) })} />
        -
        <NumInput value={state.cycle02} min={state.cycle01 + 1} max={unitMax} onChange={(v) => set({ cycle02: v })} />
        {cycleUnit}
      </Radio>
      <Radio checked={state.radio === 3} onChange={() => set({ radio: 3 })}>
        从
        <NumInput value={state.average01} min={base} max={unitMax - 1} onChange={(v) => set({ average01: v })} />
        {cycleUnit}开始，{name === 'min' ? ' ' : ''}每
        <NumInput value={state.average02} min={1} max={unitMax - state.average01} onChange={(v) => set({ average02: v })} />
        {stepLabel}
      </Radio>
      <Radio checked={state.radio === 4} onChange={() => set({ radio: 4 })}>
        指定
        <AssignSelect
          value={state.checkboxList}
          limit={ASSIGN_LIMIT[name]}
          options={assignOptions}
          onChange={(list) => set({ checkboxList: list })}
        />
      </Radio>
    </div>
  )
}

/** 日面板（对位基准 day.vue：7 个单选项含 不指定(?) / 工作日XW / 本月最后 L） */
function DayFieldPanel({ state, onSegment, onState, cron }: {
  state: DayFieldState
  cron: CronValue
  onSegment: (name: FieldName, value: string) => void
  onState: (name: FieldName, state: DayFieldState) => void
}) {
  const set = (patch: Partial<DayFieldState>) => {
    const next = { ...state, ...patch }
    onState('day', next)
    onSegment('day', buildDaySegment(next))
    // 日 → 周互斥联动（对位基准 day.vue onRadioChange 前置逻辑）
    const adjusted = weekAdjustFromDay(next.radio, cron.week)
    if (adjusted !== null) onSegment('week', adjusted)
  }
  return (
    <div style={{ padding: '12px 24px', display: 'flex', flexDirection: 'column', gap: 12 }}>
      <Radio checked={state.radio === 1} onChange={() => set({ radio: 1 })}>
        日，允许的通配符[, - * ? / L W]
      </Radio>
      <Radio checked={state.radio === 2} onChange={() => set({ radio: 2 })}>不指定</Radio>
      <Radio checked={state.radio === 3} onChange={() => set({ radio: 3 })}>
        周期从
        <NumInput value={state.cycle01} min={1} max={30}
          onChange={(v) => set({ cycle01: v, cycle02: checkNumber(state.cycle02, v + 1, 31) })} />
        -
        <NumInput value={state.cycle02} min={state.cycle01 + 1} max={31} onChange={(v) => set({ cycle02: v })} />
        日
      </Radio>
      <Radio checked={state.radio === 4} onChange={() => set({ radio: 4 })}>
        从
        <NumInput value={state.average01} min={1} max={30} onChange={(v) => set({ average01: v })} />
        号开始，每
        <NumInput value={state.average02} min={1} max={31 - state.average01} onChange={(v) => set({ average02: v })} />
        日执行一次
      </Radio>
      <Radio checked={state.radio === 5} onChange={() => set({ radio: 5 })}>
        每月
        <NumInput value={state.workday} min={1} max={31} onChange={(v) => set({ workday: v })} />
        号最近的那个工作日
      </Radio>
      <Radio checked={state.radio === 6} onChange={() => set({ radio: 6 })}>本月最后一天</Radio>
      <Radio checked={state.radio === 7} onChange={() => set({ radio: 7 })}>
        指定
        <AssignSelect
          value={state.checkboxList}
          limit={ASSIGN_LIMIT.day}
          options={seq(1, 31).map((n) => ({ label: String(n), value: n }))}
          onChange={(list) => set({ checkboxList: list })}
        />
      </Radio>
    </div>
  )
}

/** 周面板（对位基准 week.vue：6 个单选项，星期下拉 1=星期日…7=星期六；周期止点禁选 ≤ 起点，起点禁选 7） */
function WeekFieldPanel({ state, onSegment, onState, cron }: {
  state: WeekFieldState
  cron: CronValue
  onSegment: (name: FieldName, value: string) => void
  onState: (name: FieldName, state: WeekFieldState) => void
}) {
  const weekOptions = WEEK_LIST.map((w) => ({ label: w.value, value: w.key }))
  const set = (patch: Partial<WeekFieldState>) => {
    const next = { ...state, ...patch }
    onState('week', next)
    onSegment('week', buildWeekSegment(next))
    // 周 → 日互斥联动（对位基准 week.vue onRadioChange 前置逻辑）
    const adjusted = dayAdjustFromWeek(next.radio, cron.day)
    if (adjusted !== null) onSegment('day', adjusted)
  }
  return (
    <div style={{ padding: '12px 24px', display: 'flex', flexDirection: 'column', gap: 12 }}>
      <Radio checked={state.radio === 1} onChange={() => set({ radio: 1 })}>
        周，允许的通配符[, - * ? / L #]
      </Radio>
      <Radio checked={state.radio === 2} onChange={() => set({ radio: 2 })}>不指定</Radio>
      <Radio checked={state.radio === 3} onChange={() => set({ radio: 3 })}>
        周期从
        <Select size="small" style={{ width: 110 }} options={weekOptions.filter((o) => o.value !== 7)}
          value={state.cycle01} onChange={(v) => set({ cycle01: v })} />
        -
        <Select size="small" style={{ width: 110 }}
          options={weekOptions.filter((o) => o.value > state.cycle01)}
          value={state.cycle02} onChange={(v) => set({ cycle02: v })} />
      </Radio>
      <Radio checked={state.radio === 4} onChange={() => set({ radio: 4 })}>
        第
        <NumInput value={state.average01} min={1} max={4} onChange={(v) => set({ average01: v })} />
        周的
        <Select size="small" style={{ width: 110 }} options={weekOptions}
          value={state.average02} onChange={(v) => set({ average02: v })} />
      </Radio>
      <Radio checked={state.radio === 5} onChange={() => set({ radio: 5 })}>
        本月最后一个
        <Select size="small" style={{ width: 110 }} options={weekOptions}
          value={state.weekday} onChange={(v) => set({ weekday: v })} />
      </Radio>
      <Radio checked={state.radio === 6} onChange={() => set({ radio: 6 })}>
        指定
        <AssignSelect
          value={state.checkboxList}
          limit={ASSIGN_LIMIT.week}
          options={weekOptions}
          onChange={(list) => set({ checkboxList: list })}
        />
      </Radio>
    </div>
  )
}

/** 年面板（对位基准 year.vue：5 个单选项，范围 [当年, 当年+10]，步长上限 10） */
function YearFieldPanel({ state, fullYear, onSegment, onState }: {
  state: YearFieldState
  fullYear: number
  onSegment: (name: FieldName, value: string) => void
  onState: (name: FieldName, state: YearFieldState) => void
}) {
  const maxFullYear = fullYear + 10
  const set = (patch: Partial<YearFieldState>) => {
    const next = { ...state, ...patch }
    onState('year', next)
    onSegment('year', buildYearSegment(next, fullYear))
  }
  return (
    <div style={{ padding: '12px 24px', display: 'flex', flexDirection: 'column', gap: 12 }}>
      <Radio checked={state.radio === 1} onChange={() => set({ radio: 1 })}>
        不填，允许的通配符[, - * /]
      </Radio>
      <Radio checked={state.radio === 2} onChange={() => set({ radio: 2 })}>每年</Radio>
      <Radio checked={state.radio === 3} onChange={() => set({ radio: 3 })}>
        周期从
        <NumInput value={state.cycle01} min={fullYear} max={maxFullYear - 1}
          onChange={(v) => set({ cycle01: v, cycle02: checkNumber(state.cycle02, v + 1, maxFullYear) })} />
        -
        <NumInput value={state.cycle02} min={state.cycle01 + 1} max={maxFullYear} onChange={(v) => set({ cycle02: v })} />
      </Radio>
      <Radio checked={state.radio === 4} onChange={() => set({ radio: 4 })}>
        从
        <NumInput value={state.average01} min={fullYear} max={maxFullYear - 1} onChange={(v) => set({ average01: v })} />
        年开始，每
        <NumInput value={state.average02} min={1} max={10} onChange={(v) => set({ average02: v })} />
        年执行一次
      </Radio>
      <Radio checked={state.radio === 5} onChange={() => set({ radio: 5 })}>
        指定
        <AssignSelect
          value={state.checkboxList}
          limit={ASSIGN_LIMIT.year}
          options={seq(0, 8).map((i) => ({ label: String(fullYear + i), value: fullYear + i }))}
          onChange={(list) => set({ checkboxList: list })}
        />
      </Radio>
    </div>
  )
}

/** 指定多选（对位基准 el-select multiple + multiple-limit；antd maxCount 5.13+ 生效，超出即不再选中） */
function AssignSelect({ value, limit, options, onChange }: {
  value: number[]
  limit: number
  options: { label: string; value: number }[]
  onChange: (list: number[]) => void
}) {
  return (
    <Select
      mode="multiple"
      allowClear
      size="small"
      style={{ minWidth: 220 }}
      placeholder="可多选"
      maxCount={limit}
      value={value}
      options={options}
      onChange={(vals: number[]) => onChange([...new Set(vals)])}
    />
  )
}

// ---- 小工具 ----

/** InputNumber：onChange 可能收到 null（清空），基准 el-input-number 的空值同样被 checkNumber 兜底钳位，此处归一为 min */
function NumInput({ value, min, max, onChange }: {
  value: number
  min: number
  max: number
  onChange: (v: number) => void
}) {
  return (
    <InputNumber
      size="small"
      style={{ width: 64, margin: '0 4px' }}
      min={min}
      max={max}
      value={value}
      onChange={(v) => onChange(checkNumber(v ?? min, min, max))}
    />
  )
}

function seq(from: number, to: number): number[] {
  const arr: number[] = []
  for (let i = from; i <= to; i++) arr.push(i)
  return arr
}

// ---- 七域状态包（父组件 Crontab 持有） ----

export interface FieldStates {
  second: SimpleFieldState
  min: SimpleFieldState
  hour: SimpleFieldState
  day: DayFieldState
  month: SimpleFieldState
  week: WeekFieldState
  year: YearFieldState
}

/** 由表达式对象初始化全部域状态（回显；对位基准各域 changeRadioValue + ref 初值的合成效果） */
export function initFieldStates(cron: CronValue, fullYear: number): FieldStates {
  const second = simpleInit({ base: 0, unitMax: 59 })
  const min = simpleInit({ base: 0, unitMax: 59 })
  const hour = simpleInit({ base: 0, unitMax: 23 })
  const month = simpleInit({ base: 1, unitMax: 12 })
  return {
    second: parseSimpleSegment(cron.second, { base: 0, unitMax: 59 }, second),
    min: parseSimpleSegment(cron.min, { base: 0, unitMax: 59 }, min),
    hour: parseSimpleSegment(cron.hour, { base: 0, unitMax: 23 }, hour),
    month: parseSimpleSegment(cron.month, { base: 1, unitMax: 12 }, month),
    day: parseDaySegment(cron.day, {
      radio: 1, cycle01: 1, cycle02: 2, average01: 1, average02: 1, workday: 1, checkboxList: [],
    }),
    week: parseWeekSegment(cron.week, {
      radio: 2, cycle01: 2, cycle02: 3, average01: 1, average02: 2, weekday: 2, checkboxList: [],
    }),
    year: parseYearSegment(cron.year, {
      radio: 1, cycle01: fullYear, cycle02: fullYear + 1, average01: fullYear, average02: 1, checkboxList: [],
    }),
  }
}

function simpleInit(range: SimpleFieldRange): SimpleFieldState {
  return {
    radio: 1,
    cycle01: range.base,
    cycle02: range.base + 1,
    average01: range.base,
    average02: 1,
    checkboxList: [],
  }
}

// ---- 对外统一出口：按当前页签名渲染对应域面板 ----

export interface FieldPanelProps {
  name: FieldName
  cron: CronValue
  states: FieldStates
  fullYear: number
  /** 状态更新（父组件持有各域状态） */
  onState: <K extends FieldName>(name: K, state: FieldStates[K]) => void
  /** 片段更新（父组件写回表达式对象） */
  onSegment: (name: FieldName, value: string) => void
}

export default function FieldPanel({ name, cron, states, fullYear, onState, onSegment }: FieldPanelProps) {
  switch (name) {
    case 'second':
      return (
        <SimpleFieldPanel name="second" radioLabel="秒，允许的通配符[, - * /]" cycleUnit="秒" stepLabel="秒执行一次"
          state={states.second} range={{ base: 0, unitMax: 59 }} onSegment={onSegment} onState={onState} />
      )
    case 'min':
      return (
        <SimpleFieldPanel name="min" radioLabel="分钟，允许的通配符[, - * /]" cycleUnit="分钟" stepLabel="分钟执行一次"
          state={states.min} range={{ base: 0, unitMax: 59 }} onSegment={onSegment} onState={onState} />
      )
    case 'hour':
      return (
        <SimpleFieldPanel name="hour" radioLabel="小时，允许的通配符[, - * /]" cycleUnit="时" stepLabel="小时执行一次"
          state={states.hour} range={{ base: 0, unitMax: 23 }} onSegment={onSegment} onState={onState} />
      )
    case 'month':
      return (
        <SimpleFieldPanel name="month" radioLabel="月，允许的通配符[, - * /]" cycleUnit="月" stepLabel="月月执行一次"
          state={states.month} range={{ base: 1, unitMax: 12 }} onSegment={onSegment} onState={onState} />
      )
    case 'day':
      return <DayFieldPanel state={states.day} cron={cron} onSegment={onSegment} onState={onState} />
    case 'week':
      return <WeekFieldPanel state={states.week} cron={cron} onSegment={onSegment} onState={onState} />
    case 'year':
      return <YearFieldPanel state={states.year} fullYear={fullYear} onSegment={onSegment} onState={onState} />
  }
}
