// useCrud —— 6.0.0 通用 CRUD 范式（对位基准各列表页的查询/分页/多选/删除/导出模式）
// T = 行数据类型；Q = 查询参数类型

import { useCallback, useState } from 'react'
import { Modal, message } from 'antd'
import { download } from '@/utils/request'

interface UseCrudOptions<Q> {
  listApi: (query?: Record<string, unknown>) => Promise<unknown>
  delApi?: (ids: number | string | (number | string)[]) => Promise<unknown>
  exportUrl?: string
  defaultQuery: Q
}

// Q 允许任意对象形状（接口类型无索引签名也可用）
export function useCrud<T extends { [k: string]: unknown }, Q extends object>(
  opts: UseCrudOptions<Q>,
) {
  const [rows, setRows] = useState<T[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState<Q>({ ...opts.defaultQuery })
  const [ids, setIds] = useState<(number | string)[]>([])
  const [single, setSingle] = useState(false)
  const [multiple, setMultiple] = useState(false)

  const getList = useCallback(async (q?: Q) => {
    const params = { ...(query as Record<string, unknown>), ...((q || {}) as Record<string, unknown>) }
    setLoading(true)
    try {
      const res = (await opts.listApi(params)) as { rows?: T[]; total?: number }
      setRows(res.rows || [])
      setTotal(res.total || 0)
    } finally {
      setLoading(false)
    }
  }, [opts.listApi, query])

  // 查询：回第 1 页
  const handleQuery = useCallback(() => {
    void getList({ ...(opts.defaultQuery as object), pageNum: 1 } as never)
  }, [getList, opts.defaultQuery])

  // 重置表单
  const resetQuery = useCallback(() => {
    setQuery({ ...opts.defaultQuery })
    void getList(opts.defaultQuery as never)
  }, [getList, opts.defaultQuery])

  // 多选
  const handleSelectionChange = useCallback((selectedRowKeys: (number | string)[], rows: T[]) => {
    setIds(selectedRowKeys)
    setMultiple(selectedRowKeys.length > 1)
    setSingle(selectedRowKeys.length === 1)
    void rows
  }, [])

  // 删除确认（文案对齐基准：「是否确认删除编号为"xxx"的数据项？」由页面传入具体编号）
  const handleDelete = useCallback(
    (idLabel: string, idList?: (number | string)[]) => {
      const target = idList || ids
      if (!opts.delApi) return
      Modal.confirm({
        title: '系统提示',
        content: `是否确认删除编号为"${idLabel}"的数据项？`,
        okText: '确定',
        cancelText: '取消',
        onOk: async () => {
          await opts.delApi!(target)
          message.success('删除成功')
          void getList()
        },
      })
    },
    [ids, opts.delApi, getList],
  )

  // 导出（POST form-urlencoded + blob，错误确认框对齐基准）
  const handleExport = useCallback(
    (exportParams?: Record<string, unknown>) => {
      if (!opts.exportUrl) return
      Modal.confirm({
        title: '系统提示',
        content: '是否确认导出所有数据项？',
        okText: '确定',
        cancelText: '取消',
        onOk: () => download(opts.exportUrl!, { ...query, ...exportParams }, `ruoyi_${Date.now()}.xlsx`),
      })
    },
    [opts.exportUrl, query],
  )

  return {
    rows, total, loading, query, setQuery, ids, single, multiple,
    getList, handleQuery, resetQuery, handleSelectionChange, handleDelete, handleExport,
  }
}
