// DictTag —— 对位基准 components/DictTag
// 值按 separator 拆分；elTagType → antd color 映射；未匹配追加原值（showValue）

import { Tag, Typography } from 'antd'
import type { DictDataOption } from '@/store/modules/dict'

interface DictTagProps {
  options: DictDataOption[]
  value?: string | number | boolean | (string | number | boolean)[]
  showValue?: boolean
  separator?: string
}

// 基准 elTagType（Element Plus tag type）→ antd Tag color 映射
export const TAG_TYPE_COLOR: Record<string, string> = {
  default: 'default',
  '': 'default',
  success: 'green',
  info: 'blue',
  warning: 'orange',
  danger: 'red',
  primary: 'blue',
}

export default function DictTag({ options, value, showValue = true, separator = ',' }: DictTagProps) {
  if (value === undefined || value === null || value === '') return null
  const values = Array.isArray(value)
    ? value.map(String)
    : String(value).split(separator).filter((v) => v !== '')

  const matched: { label: string; color: string; cls?: string }[] = []
  const unmatch: string[] = []
  for (const v of values) {
    const hit = options.find((o) => o.value == v)
    if (hit) {
      matched.push({ label: hit.label, color: TAG_TYPE_COLOR[hit.elTagType || 'default'] || 'default', cls: hit.elTagClass })
    } else {
      unmatch.push(v)
    }
  }

  return (
    <span>
      {matched.map((m, i) => (
        <Tag key={i} color={m.color === 'default' ? undefined : m.color} className={m.cls} style={{ marginRight: 4 }}>
          {m.label}
        </Tag>
      ))}
      {showValue && unmatch.length > 0 && (
        <Typography.Text type="secondary">{unmatch.join(separator)}</Typography.Text>
      )}
    </span>
  )
}
