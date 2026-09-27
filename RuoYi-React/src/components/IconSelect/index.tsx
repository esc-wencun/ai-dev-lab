// IconSelect —— 对位基准 IconSelect（本地精灵网格 + 名称过滤 + 选中高亮）
// 菜单管理（6.0.0）消费；本批次先落组件

import { useMemo, useState } from 'react'
import { Input } from 'antd'
import SvgIcon from '@/components/SvgIcon'

interface IconSelectProps {
  activeIcon?: string
  onSelect?: (name: string) => void
}

export default function IconSelect({ activeIcon, onSelect }: IconSelectProps) {
  const [filter, setFilter] = useState('')

  // import.meta.glob 枚举精灵文件名（对位基准 requireIcons.js）
  const icons = useMemo(() => {
    const modules = import.meta.glob('../../assets/icons/svg/*.svg')
    return Object.keys(modules).map((p) =>
      p.split('/').pop()!.replace('.svg', ''),
    )
  }, [])

  const filtered = icons.filter((n) => n.includes(filter.toLowerCase()))

  return (
    <div>
      <Input
        placeholder="搜索图标名"
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
        allowClear
        style={{ marginBottom: 8 }}
      />
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 4, maxHeight: 260, overflowY: 'auto' }}>
        {filtered.map((name) => (
          <div
            key={name}
            onClick={() => onSelect?.(name)}
            style={{
              display: 'flex', flexDirection: 'column', alignItems: 'center', padding: 6,
              cursor: 'pointer', border: name === activeIcon ? '2px solid #409EFF' : '2px solid transparent',
            }}
          >
            <SvgIcon iconClass={name} size={20} />
            <span style={{ fontSize: 10, marginTop: 2, wordBreak: 'break-all' }}>{name}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
