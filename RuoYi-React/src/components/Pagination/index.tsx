// Pagination —— 对位基准 components/Pagination/index.vue（全局注册分页器，几乎所有列表页使用）
// props 对位：v-model:page → page；v-model:limit → pageSize；@pagination / update:* → onChange(page, pageSize)
// 行为对位：布局 total, sizes, prev, pager, next, jumper；切换条数后「当前页 × 新条数 > total」回第 1 页；
//          每次变更后平滑回顶（autoScroll，对位基准 scrollTo(0, 800)）
// 差异：pagerCount（<992px 视口 5 否则 7）antd 无对应 prop，页码按钮数量由 antd 自适应——纯视觉，功能等价

import { Pagination as AntdPagination } from 'antd'

interface PaginationProps {
  /** 总条数（基准必传） */
  total: number
  /** 当前页（对位 v-model:page），默认 1 */
  page?: number
  /** 每页条数（对位 v-model:limit），默认 20 */
  pageSize?: number
  /** 条数可选集，默认 [10, 20, 30, 50] */
  pageSizes?: number[]
  /** 变更后是否平滑回顶，默认 true */
  autoScroll?: boolean
  /** 是否隐藏（对位基准 hidden prop） */
  hidden?: boolean
  /** 页数/条数变化统一回调（对位 @pagination；size 变化触发的越界回 1 已在组件内修正） */
  onChange?: (page: number, pageSize: number) => void
}

export default function Pagination({
  total,
  page = 1,
  pageSize = 20,
  pageSizes = [10, 20, 30, 50],
  autoScroll = true,
  hidden = false,
  onChange,
}: PaginationProps) {
  const handleChange = (nextPage: number, nextSize: number) => {
    // 对位基准 handleSizeChange：切换条数后当前页 × 新条数超过 total 则回第 1 页（total=0 时同样回 1）
    let target = nextPage
    if (nextSize !== pageSize && page * nextSize > total) {
      target = 1
    }
    onChange?.(target, nextSize)
    if (autoScroll) {
      // 对位基准 scrollTo(0, 800)：平滑回顶
      window.scrollTo({ top: 0, behavior: 'smooth' })
    }
  }

  if (hidden) return null

  return (
    <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 20 }}>
      <AntdPagination
        current={page}
        pageSize={pageSize}
        total={total}
        showSizeChanger
        showQuickJumper
        pageSizeOptions={pageSizes.map((s) => String(s))}
        showTotal={(t) => `共 ${t} 条`}
        onChange={handleChange}
      />
    </div>
  )
}
