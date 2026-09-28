// RightToolbar —— 对位基准 components/RightToolbar/index.vue（列表页右上角工具条）
// 三个圆钮：搜索显隐切换（对位 v-model:showSearch）+ 刷新（对位 @queryTable）+ 列显隐下拉（对位 columns）
// 差异（对齐 deviations #4）：
//   1. 基准的搜索折叠动画在组件内向上查找 .el-form 做 max-height/opacity 过渡——纯视觉糖，简化为
//      onShowSearchChange 回调 + 页面侧 display 切换（v-show 等价），不做 DOM 上溯与动画；
//   2. 显隐列仅实现基准默认的 checkbox 下拉形态；transfer 穿梭框形态无页面消费，不实现；
//   3. storageKey 记忆按列 key 存 localStorage（基准对象格式同款），数组下标格式不实现。

import { useEffect, useRef } from 'react'
import { Button, Checkbox, Divider, Popover, Tooltip } from 'antd'
import { MenuOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import cache from '@/plugins/cache'

/** 列显隐信息项（对位基准数组格式项 {label,key,visible}） */
export interface ToolbarColumn {
  label: string
  key: string
  visible: boolean
}

interface RightToolbarProps {
  /** 搜索表单显隐（受控，对位 v-model:showSearch），默认 true */
  showSearch?: boolean
  /** 搜索显隐回调（对位 update:showSearch） */
  onShowSearchChange?: (show: boolean) => void
  /** 是否显示检索切换钮，默认 true */
  search?: boolean
  /** 列显隐信息（传入才显示显隐列按钮）；由页面持有 state，变更经 onColumnsChange 回流 */
  columns?: ToolbarColumn[]
  /** 列显隐变更回调 */
  onColumnsChange?: (columns: ToolbarColumn[]) => void
  /** 列显隐状态记忆的 localStorage key（传入则启用记忆，挂载时恢复） */
  storageKey?: string
  /** 刷新回调（对位 @queryTable） */
  onRefresh?: () => void
  /** 右外边距(px)，默认 10 */
  gutter?: number
}

export default function RightToolbar({
  showSearch = true,
  onShowSearchChange,
  search = true,
  columns,
  onColumnsChange,
  storageKey,
  onRefresh,
  gutter = 10,
}: RightToolbarProps) {
  // 挂载时从 localStorage 恢复列显隐状态（对位基准 setup 内同步恢复；ref 防重复应用）
  const restoredRef = useRef(false)
  useEffect(() => {
    if (!columns || !storageKey || restoredRef.current) return
    restoredRef.current = true
    try {
      const saved = cache.local.getJSON<Record<string, boolean>>(storageKey)
      if (saved && typeof saved === 'object') {
        onColumnsChange?.(columns.map((c) => (c.key in saved ? { ...c, visible: !!saved[c.key] } : c)))
      }
    } catch {
      // 恢复失败静默忽略（对位基准 try/catch）
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 持久化当前列显隐状态到 localStorage（按列 key）
  const saveStorage = (next: ToolbarColumn[]) => {
    if (!storageKey) return
    try {
      const state: Record<string, boolean> = {}
      next.forEach((c) => {
        state[c.key] = c.visible
      })
      cache.local.setJSON(storageKey, state)
    } catch {
      // 忽略
    }
  }

  const setColVisible = (key: string, visible: boolean) => {
    if (!columns) return
    const next = columns.map((c) => (c.key === key ? { ...c, visible } : c))
    onColumnsChange?.(next)
    saveStorage(next)
  }

  const toggleCheckAll = () => {
    if (!columns) return
    const target = !columns.every((c) => c.visible)
    const next = columns.map((c) => ({ ...c, visible: target }))
    onColumnsChange?.(next)
    saveStorage(next)
  }

  const checkedCount = columns?.filter((c) => c.visible).length ?? 0
  const allChecked = !!columns && checkedCount === columns.length
  const indeterminate = checkedCount > 0 && !allChecked

  return (
    // 对位基准 .top-right-btn 全局样式（margin-left: auto 右对齐 + 右外边距）
    <div
      className="top-right-btn"
      style={{ marginLeft: 'auto', marginRight: gutter ? gutter / 2 : undefined, flexShrink: 0 }}
    >
      {search && (
        <Tooltip title={showSearch ? '隐藏搜索' : '显示搜索'}>
          <Button
            shape="circle"
            icon={<SearchOutlined />}
            onClick={() => onShowSearchChange?.(!showSearch)}
          />
        </Tooltip>
      )}
      <Tooltip title="刷新">
        <Button shape="circle" icon={<ReloadOutlined />} onClick={() => onRefresh?.()} style={{ marginLeft: 8 }} />
      </Tooltip>
      {columns && columns.length > 0 && (
        <Popover
          trigger="click"
          content={
            <div style={{ minWidth: 120 }}>
              <Checkbox indeterminate={indeterminate} checked={allChecked} onChange={toggleCheckAll}>
                列展示
              </Checkbox>
              <Divider style={{ margin: '4px 0' }} />
              {columns.map((col) => (
                <div key={col.key} style={{ padding: '2px 0', lineHeight: '30px' }}>
                  <Checkbox checked={col.visible} onChange={(e) => setColVisible(col.key, e.target.checked)}>
                    {col.label}
                  </Checkbox>
                </div>
              ))}
            </div>
          }
        >
          <Tooltip title="显隐列">
            <Button shape="circle" icon={<MenuOutlined />} style={{ marginLeft: 8 }} />
          </Tooltip>
        </Popover>
      )}
    </div>
  )
}
