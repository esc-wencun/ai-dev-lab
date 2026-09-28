// Crontab 组件 —— 对位基准 RuoYi-Vue3/src/components/Crontab/index.vue
// 消费方式对位 job 页：expression prop 回显 + fill 回调（确定）/ hide 回调（取消/确定后关弹窗）。
// 结构：七域页签（FieldPanel）+ 时间表达式表 + 最近5次运行时间（result.ts 纯算法）+ 确定/重置/取消。
// 回显对位基准 resolveExp：split(/\s+/) 后 >=6 段才解析，否则维持默认值（解析不了的行为也对位）。
// hour 隐藏副作用对位：hour 片段变化时把仍为 * 的分/秒归一为 '0'（见 hourSegmentSideEffect）。

import { useMemo, useState } from 'react'
import { Button, Tabs, Tooltip } from 'antd'
import {
  buildCronExpression,
  parseCronExpression,
  DEFAULT_CRON,
  type CronValue,
} from './expression'
import { getNextRuns } from './result'
import FieldPanel, { initFieldStates, type FieldName, type FieldStates } from './FieldPanel'

export interface CrontabProps {
  /** 外部传入的表达式（回显） */
  expression?: string
  /** 选定表达式回调（对位基准 fill 事件，参数为拼接后的完整表达式） */
  onFill?: (expression: string) => void
  /** 关闭回调（对位基准 hide 事件；基准「确定」= fill + hide，此处由父组件在 onFill 里关闭） */
  onHide?: () => void
}

const TAB_ITEMS: { key: FieldName; label: string }[] = [
  { key: 'second', label: '秒' },
  { key: 'min', label: '分钟' },
  { key: 'hour', label: '小时' },
  { key: 'day', label: '日' },
  { key: 'month', label: '月' },
  { key: 'week', label: '周' },
  { key: 'year', label: '年' },
]

const TAB_TITLES = ['秒', '分钟', '小时', '日', '月', '周', '年']

export default function Crontab({ expression, onFill, onHide }: CrontabProps) {
  const fullYear = useMemo(() => new Date().getFullYear(), [])
  // 表达式对象（对位基准 crontabValueObj）；初值 = 默认值回显外部 expression（>=6 段才解析）
  const [cron, setCron] = useState<CronValue>(() => parseCronExpression(expression ?? '') ?? { ...DEFAULT_CRON })
  // 各域 UI 状态（对位基准各域 vue 的 ref 状态；组件随弹窗 destroyOnHidden 重挂载，挂载时初始化即等价基准的 watch 回显）
  const [states, setStates] = useState<FieldStates>(() => initFieldStates(
    parseCronExpression(expression ?? '') ?? { ...DEFAULT_CRON },
    fullYear,
  ))
  const [activeTab, setActiveTab] = useState<FieldName>('second')

  const crontabValueString = buildCronExpression(cron)
  // 最近 5 次运行时间（纯算法同步计算；基准 isShow/「计算结果中...」占位在同一 tick 内实际不会渲染，行为一致）
  const nextRuns = useMemo(() => {
    try {
      return getNextRuns(crontabValueString)
    } catch {
      return ['表达式无法解析！']
    }
  }, [crontabValueString])

  /** 片段更新（对位基准 updateCrontabValue）+ hour 域隐藏副作用 */
  const handleSegment = (name: FieldName, value: string) => {
    setCron((prev) => {
      const next = { ...prev, [name]: value }
      // 对位基准 hour.vue changeRadioValue 前置副作用：hour 片段变化时把仍为 * 的分/秒归一为 0
      if (name === 'hour') {
        if (next.min === '*') next.min = '0'
        if (next.second === '*') next.second = '0'
      }
      return next
    })
  }

  const handleState = <K extends FieldName>(name: K, state: FieldStates[K]) => {
    setStates((prev) => ({ ...prev, [name]: state }))
  }

  /** 重置（对位基准 clearCron：还原默认值；各域 UI 状态一并还原为初值） */
  const handleReset = () => {
    setCron({ ...DEFAULT_CRON })
    setStates(initFieldStates({ ...DEFAULT_CRON }, fullYear))
  }

  // 时间表达式表：七域 + 完整表达式，超长走 Tooltip（对位基准 length<10 / <90 的展示分界）
  const cells: (string | undefined)[] = [
    cron.second, cron.min, cron.hour, cron.day, cron.month, cron.week, cron.year,
    crontabValueString,
  ]
  const short = [10, 10, 10, 10, 10, 10, 10, 90]

  return (
    <div>
      <Tabs
        type="card"
        activeKey={activeTab}
        onChange={(k) => setActiveTab(k as FieldName)}
        items={TAB_ITEMS.map((t) => ({
          key: t.key,
          label: t.label,
          children: (
            <FieldPanel name={t.key} cron={cron} states={states} fullYear={fullYear}
              onState={handleState} onSegment={handleSegment} />
          ),
        }))}
      />

      <div style={{ margin: '10px auto', fontSize: 12 }}>
        <div style={{ textAlign: 'center', fontWeight: 600, marginBottom: 4 }}>时间表达式</div>
        <table style={{ width: '100%', textAlign: 'center', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              {TAB_TITLES.map((t) => <th key={t}>{t}</th>)}
              <th>Cron 表达式</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              {cells.map((c, i) => {
                const span = (
                  <span style={{
                    display: 'block', border: '1px solid #e8e8e8', lineHeight: '30px',
                    whiteSpace: 'nowrap', overflow: 'hidden',
                  }}>{c}</span>
                )
                return (
                  <td key={i} style={{ padding: 2, maxWidth: i === cells.length - 1 ? 320 : 72 }}>
                    {(c ?? '').length < short[i] ? span : <Tooltip title={c}>{span}</Tooltip>}
                  </td>
                )
              })}
            </tr>
          </tbody>
        </table>
      </div>

      <div style={{ border: '1px solid #e8e8e8', padding: '8px 12px', margin: '10px auto' }}>
        <div style={{ textAlign: 'center', fontWeight: 600, marginBottom: 4 }}>最近5次运行时间</div>
        <ul style={{ margin: 0, paddingLeft: 24, maxHeight: 160, overflowY: 'auto', fontSize: 12, lineHeight: '24px' }}>
          {nextRuns.map((item, i) => <li key={i}>{item}</li>)}
        </ul>
      </div>

      <div style={{ textAlign: 'center', marginTop: 16 }}>
        <Button type="primary" style={{ marginRight: 8 }} onClick={() => onFill?.(crontabValueString)}>确定</Button>
        <Button type="default" danger style={{ marginRight: 8 }} onClick={handleReset}>重置</Button>
        <Button onClick={onHide}>取消</Button>
      </div>
    </div>
  )
}
